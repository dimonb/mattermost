// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/gorilla/mux"
	svg "github.com/h2non/go-is-svg"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/app/request"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/product"
	"github.com/mattermost/mattermost-server/v6/services/marketplace"
	"github.com/mattermost/mattermost-server/v6/shared/filestore"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/mattermost/mattermost-server/v6/utils/fileutils"
)

const prepackagedPluginsDir = "prepackaged_plugins"

type pluginSignaturePath struct {
	pluginID      string
	path          string
	signaturePath string
}

// Ensure routerService implements `product.RouterService`
var _ product.RouterService = (*routerService)(nil)

type routerService struct {
	mu        sync.Mutex
	routerMap map[string]*mux.Router
}

func newRouterService() *routerService {
	return &routerService{
		routerMap: make(map[string]*mux.Router),
	}
}

func (rs *routerService) RegisterRouter(productID string, sub *mux.Router) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.routerMap[productID] = sub
}

func (rs *routerService) getHandler(productID string) (http.Handler, bool) {
	handler, ok := rs.routerMap[productID]
	return handler, ok
}

// GetPluginsEnvironment returns the plugin environment for use if plugins are enabled and
// initialized.
//
// To get the plugins environment when the plugins are disabled, manually acquire the plugins
// lock instead.
func (ch *Channels) GetPluginsEnvironment() *plugin.Environment {
	if !*ch.cfgSvc.Config().PluginSettings.Enable {
		return nil
	}

	ch.pluginsLock.RLock()
	defer ch.pluginsLock.RUnlock()

	return ch.pluginsEnvironment
}

// GetPluginsEnvironment returns the plugin environment for use if plugins are enabled and
// initialized.
//
// To get the plugins environment when the plugins are disabled, manually acquire the plugins
// lock instead.
func (a *App) GetPluginsEnvironment() *plugin.Environment {
	return a.ch.GetPluginsEnvironment()
}

func (ch *Channels) SetPluginsEnvironment(pluginsEnvironment *plugin.Environment) {
	ch.pluginsLock.Lock()
	defer ch.pluginsLock.Unlock()

	ch.pluginsEnvironment = pluginsEnvironment
}

func (ch *Channels) syncPluginsActiveState(c request.CTX) {
	// Acquiring lock manually, as plugins might be disabled. See GetPluginsEnvironment.
	ch.pluginsLock.RLock()
	pluginsEnvironment := ch.pluginsEnvironment
	ch.pluginsLock.RUnlock()

	if pluginsEnvironment == nil {
		return
	}

	config := ch.cfgSvc.Config().PluginSettings

	if *config.Enable {
		availablePlugins, err := pluginsEnvironment.Available()
		if err != nil {
			ch.srv.Log().Error("Unable to get available plugins", mlog.Err(err))
			return
		}

		// Determine which plugins need to be activated or deactivated.
		disabledPlugins := []*model.BundleInfo{}
		enabledPlugins := []*model.BundleInfo{}
		for _, plugin := range availablePlugins {
			pluginID := plugin.Manifest.Id
			pluginEnabled := false
			if state, ok := config.PluginStates[pluginID]; ok {
				pluginEnabled = state.Enable
			}

			if hasOverride, value := ch.getPluginStateOverride(pluginID); hasOverride {
				pluginEnabled = value
			}

			if pluginEnabled {
				enabledPlugins = append(enabledPlugins, plugin)
			} else {
				disabledPlugins = append(disabledPlugins, plugin)
			}
		}

		// Concurrently activate/deactivate each plugin appropriately.
		var wg sync.WaitGroup

		// Deactivate any plugins that have been disabled.
		for _, plugin := range disabledPlugins {
			wg.Add(1)
			go func(plugin *model.BundleInfo) {
				defer wg.Done()

				deactivated := pluginsEnvironment.Deactivate(plugin.Manifest.Id)
				if deactivated && plugin.Manifest.HasClient() {
					message := model.NewWebSocketEvent(model.WebsocketEventPluginDisabled, "", "", "", nil, "")
					message.Add("manifest", plugin.Manifest.ClientManifest())
					ch.srv.Publish(c, message)
				}
			}(plugin)
		}

		// Activate any plugins that have been enabled
		for _, plugin := range enabledPlugins {
			wg.Add(1)
			go func(plugin *model.BundleInfo) {
				defer wg.Done()

				pluginID := plugin.Manifest.Id
				updatedManifest, activated, err := pluginsEnvironment.Activate(pluginID)
				if err != nil {
					plugin.WrapLogger(ch.srv.Log()).Error("Unable to activate plugin", mlog.Err(err))
					return
				}

				if activated {
					// Notify all cluster clients if ready
					if err := ch.notifyPluginEnabled(c, updatedManifest); err != nil {
						ch.srv.Log().Error("Failed to notify cluster on plugin enable", mlog.Err(err))
					}
				}
			}(plugin)
		}
		wg.Wait()
	} else { // If plugins are disabled, shutdown plugins.
		pluginsEnvironment.Shutdown()
	}

	if err := ch.notifyPluginStatusesChanged(c); err != nil {
		c.Logger().Warn("failed to notify plugin status changed", mlog.Err(err))
	}

	if err := ch.notifyIntegrationsUsageChanged(c); err != nil {
		c.Logger().Warn("Failed to notify integrations usage changed", mlog.Err(err))
	}
}

func (a *App) NewPluginAPI(c request.CTX, manifest *model.Manifest) plugin.API {
	return NewPluginAPI(a, c, manifest)
}

func (a *App) InitPlugins(c *request.Context, pluginDir, webappPluginDir string) {
	a.ch.initPlugins(c, pluginDir, webappPluginDir)
}

func (ch *Channels) initPlugins(c request.CTX, pluginDir, webappPluginDir string) {
	// Acquiring lock manually, as plugins might be disabled. See GetPluginsEnvironment.
	ch.pluginsLock.RLock()
	pluginsEnvironment := ch.pluginsEnvironment
	ch.pluginsLock.RUnlock()
	if pluginsEnvironment != nil || !*ch.cfgSvc.Config().PluginSettings.Enable {
		ch.syncPluginsActiveState(c)
		if pluginsEnvironment != nil {
			pluginsEnvironment.TogglePluginHealthCheckJob(*ch.cfgSvc.Config().PluginSettings.EnableHealthCheck)
		}
		return
	}

	ch.srv.Log().Info("Starting up plugins")

	if err := os.Mkdir(pluginDir, 0744); err != nil && !os.IsExist(err) {
		c.Logger().Error("Failed to start up plugins", mlog.Err(err))
		return
	}

	if err := os.Mkdir(webappPluginDir, 0744); err != nil && !os.IsExist(err) {
		c.Logger().Error("Failed to start up plugins", mlog.Err(err))
		return
	}

	newAPIFunc := func(manifest *model.Manifest) plugin.API {
		return New(ServerConnector(ch)).NewPluginAPI(c, manifest)
	}

	env, err := plugin.NewEnvironment(newAPIFunc, NewDriverImpl(ch.srv), pluginDir, webappPluginDir, ch.srv.Log(), ch.srv.GetMetrics())
	if err != nil {
		c.Logger().Error("Failed to start up plugins", mlog.Err(err))
		return
	}
	ch.pluginsLock.Lock()
	ch.pluginsEnvironment = env
	ch.pluginsLock.Unlock()

	ch.pluginsEnvironment.TogglePluginHealthCheckJob(*ch.cfgSvc.Config().PluginSettings.EnableHealthCheck)

	if err := ch.syncPlugins(c); err != nil {
		c.Logger().Error("Failed to sync plugins from the file store", mlog.Err(err))
	}

	plugins := ch.processPrepackagedPlugins(c, prepackagedPluginsDir)
	pluginsEnvironment = ch.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		c.Logger().Info("Plugins environment not found, server is likely shutting down")
		return
	}
	pluginsEnvironment.SetPrepackagedPlugins(plugins)

	ch.installFeatureFlagPlugins(c)

	// Sync plugin active state when config changes. Also notify plugins.
	ch.pluginsLock.Lock()
	ch.RemoveConfigListener(ch.pluginConfigListenerID)
	ch.pluginConfigListenerID = ch.AddConfigListener(func(old, new *model.Config) {
		// If plugin status remains unchanged, only then run this.
		// Because (*App).InitPlugins is already run as a config change hook.
		if *old.PluginSettings.Enable == *new.PluginSettings.Enable {
			ch.installFeatureFlagPlugins(c)
			ch.syncPluginsActiveState(c)
		}
		if pluginsEnvironment := ch.GetPluginsEnvironment(); pluginsEnvironment != nil {
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				if err := hooks.OnConfigurationChange(); err != nil {
					ch.srv.Log().Error("Plugin OnConfigurationChange hook failed", mlog.Err(err))
				}
				return true
			}, plugin.OnConfigurationChangeID)
		}
	})
	ch.pluginsLock.Unlock()

	ch.syncPluginsActiveState(c)
}

// SyncPlugins synchronizes the plugins installed locally
// with the plugin bundles available in the file store.
func (a *App) SyncPlugins(c request.CTX) *model.AppError {
	return a.ch.syncPlugins(c)
}

// SyncPlugins synchronizes the plugins installed locally
// with the plugin bundles available in the file store.
func (ch *Channels) syncPlugins(c request.CTX) *model.AppError {
	c.Logger().Info("Syncing plugins from the file store")

	pluginsEnvironment := ch.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return model.NewAppError("SyncPlugins", "app.plugin.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	availablePlugins, err := pluginsEnvironment.Available()
	if err != nil {
		return model.NewAppError("SyncPlugins", "app.plugin.sync.read_local_folder.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	var wg sync.WaitGroup
	for _, plugin := range availablePlugins {
		wg.Add(1)
		go func(pluginID string) {
			defer wg.Done()
			// Only handle managed plugins with .filestore flag file.
			_, err := os.Stat(filepath.Join(*ch.cfgSvc.Config().PluginSettings.Directory, pluginID, managedPluginFileName))
			if os.IsNotExist(err) {
				c.Logger().Warn("Skipping sync for unmanaged plugin", mlog.String("plugin_id", pluginID))
			} else if err != nil {
				c.Logger().Error("Skipping sync for plugin after failure to check if managed", mlog.String("plugin_id", pluginID), mlog.Err(err))
			} else {
				c.Logger().Debug("Removing local installation of managed plugin before sync", mlog.String("plugin_id", pluginID))
				if err := ch.removePluginLocally(pluginID); err != nil {
					c.Logger().Error("Failed to remove local installation of managed plugin before sync", mlog.String("plugin_id", pluginID), mlog.Err(err))
				}
			}
		}(plugin.Manifest.Id)
	}
	wg.Wait()

	// Install plugins from the file store.
	pluginSignaturePathMap, appErr := ch.getPluginsFromFolder(c)
	if appErr != nil {
		return appErr
	}

	for _, plugin := range pluginSignaturePathMap {
		wg.Add(1)
		go func(plugin *pluginSignaturePath) {
			defer wg.Done()
			reader, appErr := ch.srv.fileReader(plugin.path)
			if appErr != nil {
				c.Logger().Error("Failed to open plugin bundle from file store.", mlog.String("bundle", plugin.path), mlog.Err(appErr))
				return
			}
			defer reader.Close()

			var signature filestore.ReadCloseSeeker
			if *ch.cfgSvc.Config().PluginSettings.RequirePluginSignature {
				signature, appErr = ch.srv.fileReader(plugin.signaturePath)
				if appErr != nil {
					c.Logger().Error("Failed to open plugin signature from file store.", mlog.Err(appErr))
					return
				}
				defer signature.Close()
			}

			c.Logger().Info("Syncing plugin from file store", mlog.String("bundle", plugin.path))
			if _, err := ch.installPluginLocally(c, reader, signature, installPluginLocallyAlways); err != nil {
				c.Logger().Error("Failed to sync plugin from file store", mlog.String("bundle", plugin.path), mlog.Err(err))
			}
		}(plugin)
	}

	wg.Wait()
	return nil
}

func (ch *Channels) ShutDownPlugins(c request.CTX) {
	// Acquiring lock manually, as plugins might be disabled. See GetPluginsEnvironment.
	ch.pluginsLock.RLock()
	pluginsEnvironment := ch.pluginsEnvironment
	ch.pluginsLock.RUnlock()
	if pluginsEnvironment == nil {
		return
	}

	c.Logger().Info("Shutting down plugins")

	pluginsEnvironment.Shutdown()

	ch.RemoveConfigListener(ch.pluginConfigListenerID)
	ch.pluginConfigListenerID = ""

	// Acquiring lock manually before cleaning up PluginsEnvironment.
	ch.pluginsLock.Lock()
	defer ch.pluginsLock.Unlock()
	if ch.pluginsEnvironment == pluginsEnvironment {
		ch.pluginsEnvironment = nil
	} else {
		c.Logger().Warn("Another PluginsEnvironment detected while shutting down plugins.")
	}
}

func (a *App) GetActivePluginManifests() ([]*model.Manifest, *model.AppError) {
	pluginsEnvironment := a.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return nil, model.NewAppError("GetActivePluginManifests", "app.plugin.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	plugins := pluginsEnvironment.Active()

	manifests := make([]*model.Manifest, len(plugins))
	for i, plugin := range plugins {
		manifests[i] = plugin.Manifest
	}

	return manifests, nil
}

// EnablePlugin will set the config for an installed plugin to enabled, triggering asynchronous
// activation if inactive anywhere in the cluster.
// Notifies cluster peers through config change.
func (a *App) EnablePlugin(c request.CTX, id string) *model.AppError {
	appErr := a.checkIfIntegrationsMeetFreemiumLimits(c, []string{id})
	if appErr != nil {
		return appErr
	}

	return a.ch.enablePlugin(id)
}

func (ch *Channels) enablePlugin(id string) *model.AppError {
	pluginsEnvironment := ch.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return model.NewAppError("EnablePlugin", "app.plugin.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	availablePlugins, err := pluginsEnvironment.Available()
	if err != nil {
		return model.NewAppError("EnablePlugin", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	id = strings.ToLower(id)

	var manifest *model.Manifest
	for _, p := range availablePlugins {
		if p.Manifest.Id == id {
			manifest = p.Manifest
			break
		}
	}

	if manifest == nil {
		return model.NewAppError("EnablePlugin", "app.plugin.not_installed.app_error", nil, "", http.StatusNotFound)
	}

	ch.cfgSvc.UpdateConfig(func(cfg *model.Config) {
		cfg.PluginSettings.PluginStates[id] = &model.PluginState{Enable: true}
	})

	// This call will implicitly invoke SyncPluginsActiveState which will activate enabled plugins.
	if _, _, err := ch.cfgSvc.SaveConfig(ch.cfgSvc.Config(), true); err != nil {
		if err.Id == "ent.cluster.save_config.error" {
			return model.NewAppError("EnablePlugin", "app.plugin.cluster.save_config.app_error", nil, "", http.StatusInternalServerError)
		}
		return model.NewAppError("EnablePlugin", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	return nil
}

// DisablePlugin will set the config for an installed plugin to disabled, triggering deactivation if active.
// Notifies cluster peers through config change.
func (a *App) DisablePlugin(id string) *model.AppError {
	appErr := a.ch.disablePlugin(id)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (ch *Channels) disablePlugin(id string) *model.AppError {
	pluginsEnvironment := ch.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return model.NewAppError("DisablePlugin", "app.plugin.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	availablePlugins, err := pluginsEnvironment.Available()
	if err != nil {
		return model.NewAppError("DisablePlugin", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	id = strings.ToLower(id)

	var manifest *model.Manifest
	for _, p := range availablePlugins {
		if p.Manifest.Id == id {
			manifest = p.Manifest
			break
		}
	}

	if manifest == nil {
		return model.NewAppError("DisablePlugin", "app.plugin.not_installed.app_error", nil, "", http.StatusNotFound)
	}

	ch.cfgSvc.UpdateConfig(func(cfg *model.Config) {
		cfg.PluginSettings.PluginStates[id] = &model.PluginState{Enable: false}
	})
	ch.unregisterPluginCommands(id)

	// This call will implicitly invoke SyncPluginsActiveState which will deactivate disabled plugins.
	if _, _, err := ch.cfgSvc.SaveConfig(ch.cfgSvc.Config(), true); err != nil {
		return model.NewAppError("DisablePlugin", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	return nil
}

func (ch *Channels) notifyIntegrationsUsageChanged(c request.CTX) *model.AppError {
	usage, appErr := ch.getIntegrationsUsage()
	if appErr != nil {
		return appErr
	}

	message := model.NewWebSocketEvent(model.WebsocketEventIntegrationsUsageChanged, "", "", "", nil, "")
	message.Add("usage", usage)
	message.GetBroadcast().ContainsSensitiveData = true
	ch.Publish(c, message)

	return nil
}

func (a *App) GetPlugins() (*model.PluginsResponse, *model.AppError) {
	pluginsEnvironment := a.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return nil, model.NewAppError("GetPlugins", "app.plugin.disabled.app_error", nil, "", http.StatusNotImplemented)
	}

	availablePlugins, err := pluginsEnvironment.Available()
	if err != nil {
		return nil, model.NewAppError("GetPlugins", "app.plugin.get_plugins.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}
	resp := &model.PluginsResponse{Active: []*model.PluginInfo{}, Inactive: []*model.PluginInfo{}}
	for _, plugin := range availablePlugins {
		if plugin.Manifest == nil {
			continue
		}

		info := &model.PluginInfo{
			Manifest: *plugin.Manifest,
		}

		if pluginsEnvironment.IsActive(plugin.Manifest.Id) {
			resp.Active = append(resp.Active, info)
		} else {
			resp.Inactive = append(resp.Inactive, info)
		}
	}

	return resp, nil
}

// GetMarketplacePlugins returns a list of plugins from the marketplace-server,
// and plugins that are installed locally.
func (a *App) GetMarketplacePlugins(c request.CTX, filter *model.MarketplacePluginFilter) ([]*model.MarketplacePlugin, *model.AppError) {
	plugins := map[string]*model.MarketplacePlugin{}

	if *a.Config().PluginSettings.EnableRemoteMarketplace && !filter.LocalOnly {
		p, appErr := a.getRemotePlugins(c)
		if appErr != nil {
			return nil, appErr
		}
		plugins = p
	}

	// Some plugin don't work on cloud. The remote Marketplace is aware of this fact,
	// but prepackaged plugins are not. Hence, on a cloud installation prepackaged plugins
	// shouldn't be shown in the Marketplace modal.
	// This is a short term fix. The long term solution is to have a separate set of
	// prepacked plugins for cloud: https://mattermost.atlassian.net/browse/MM-31331.
	license := a.Srv().License(c)
	if license == nil || !*license.Features.Cloud {
		appErr := a.mergePrepackagedPlugins(plugins)
		if appErr != nil {
			return nil, appErr
		}
	}

	appErr := a.mergeLocalPlugins(c, plugins)
	if appErr != nil {
		return nil, appErr
	}

	// Filter plugins.
	var result []*model.MarketplacePlugin
	for _, p := range plugins {
		if pluginMatchesFilter(p.Manifest, filter.Filter) {
			result = append(result, p)
		}
	}

	// Sort result alphabetically.
	sort.SliceStable(result, func(i, j int) bool {
		return strings.ToLower(result[i].Manifest.Name) < strings.ToLower(result[j].Manifest.Name)
	})

	return result, nil
}

// getPrepackagedPlugin returns a pre-packaged plugin.
//
// If version is empty, the first matching plugin is returned.
func (ch *Channels) getPrepackagedPlugin(pluginID, version string) (*plugin.PrepackagedPlugin, *model.AppError) {
	pluginsEnvironment := ch.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return nil, model.NewAppError("getPrepackagedPlugin", "app.plugin.config.app_error", nil, "plugin environment is nil", http.StatusInternalServerError)
	}

	prepackagedPlugins := pluginsEnvironment.PrepackagedPlugins()
	for _, p := range prepackagedPlugins {
		if p.Manifest.Id == pluginID && (version == "" || p.Manifest.Version == version) {
			return p, nil
		}
	}

	return nil, model.NewAppError("getPrepackagedPlugin", "app.plugin.marketplace_plugins.not_found.app_error", nil, "", http.StatusInternalServerError)
}

// getRemoteMarketplacePlugin returns plugin from marketplace-server.
//
// If version is empty, the latest compatible version is used.
func (ch *Channels) getRemoteMarketplacePlugin(c request.CTX, pluginID, version string) (*model.BaseMarketplacePlugin, *model.AppError) {
	marketplaceClient, err := marketplace.NewClient(
		*ch.cfgSvc.Config().PluginSettings.MarketplaceURL,
		ch.srv.HTTPService(),
	)
	if err != nil {
		return nil, model.NewAppError("GetMarketplacePlugin", "app.plugin.marketplace_client.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	filter := ch.getBaseMarketplaceFilter(c)
	filter.PluginId = pluginID

	var plugin *model.BaseMarketplacePlugin
	if version != "" {
		plugin, err = marketplaceClient.GetPlugin(filter, version)
	} else {
		plugin, err = marketplaceClient.GetLatestPlugin(filter)
	}
	if err != nil {
		return nil, model.NewAppError("GetMarketplacePlugin", "app.plugin.marketplace_plugins.not_found.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	return plugin, nil
}

func (a *App) getRemotePlugins(c request.CTX) (map[string]*model.MarketplacePlugin, *model.AppError) {
	result := map[string]*model.MarketplacePlugin{}

	pluginsEnvironment := a.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return nil, model.NewAppError("getRemotePlugins", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError)
	}

	marketplaceClient, err := marketplace.NewClient(
		*a.Config().PluginSettings.MarketplaceURL,
		a.HTTPService(),
	)
	if err != nil {
		return nil, model.NewAppError("getRemotePlugins", "app.plugin.marketplace_client.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	filter := a.getBaseMarketplaceFilter(c)
	// Fetch all plugins from marketplace.
	filter.PerPage = -1

	marketplacePlugins, err := marketplaceClient.GetPlugins(filter)
	if err != nil {
		return nil, model.NewAppError("getRemotePlugins", "app.plugin.marketplace_client.failed_to_fetch", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	for _, p := range marketplacePlugins {
		if p.Manifest == nil {
			continue
		}

		result[p.Manifest.Id] = &model.MarketplacePlugin{BaseMarketplacePlugin: p}
	}

	return result, nil
}

// mergePrepackagedPlugins merges pre-packaged plugins to remote marketplace plugins list.
func (a *App) mergePrepackagedPlugins(remoteMarketplacePlugins map[string]*model.MarketplacePlugin) *model.AppError {
	pluginsEnvironment := a.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return model.NewAppError("mergePrepackagedPlugins", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError)
	}

	for _, prepackaged := range pluginsEnvironment.PrepackagedPlugins() {
		if prepackaged.Manifest == nil {
			continue
		}

		prepackagedMarketplace := &model.MarketplacePlugin{
			BaseMarketplacePlugin: &model.BaseMarketplacePlugin{
				HomepageURL:     prepackaged.Manifest.HomepageURL,
				IconData:        prepackaged.IconData,
				ReleaseNotesURL: prepackaged.Manifest.ReleaseNotesURL,
				Manifest:        prepackaged.Manifest,
			},
		}

		// If not available in marketplace, add the prepackaged
		if remoteMarketplacePlugins[prepackaged.Manifest.Id] == nil {
			remoteMarketplacePlugins[prepackaged.Manifest.Id] = prepackagedMarketplace
			continue
		}

		// If available in the marketplace, only overwrite if newer.
		prepackagedVersion, err := semver.Parse(prepackaged.Manifest.Version)
		if err != nil {
			return model.NewAppError("mergePrepackagedPlugins", "app.plugin.invalid_version.app_error", nil, "", http.StatusBadRequest).Wrap(err)
		}

		marketplacePlugin := remoteMarketplacePlugins[prepackaged.Manifest.Id]
		marketplaceVersion, err := semver.Parse(marketplacePlugin.Manifest.Version)
		if err != nil {
			return model.NewAppError("mergePrepackagedPlugins", "app.plugin.invalid_version.app_error", nil, "", http.StatusBadRequest).Wrap(err)
		}

		if prepackagedVersion.GT(marketplaceVersion) {
			remoteMarketplacePlugins[prepackaged.Manifest.Id] = prepackagedMarketplace
		}
	}

	return nil
}

// mergeLocalPlugins merges locally installed plugins to remote marketplace plugins list.
func (a *App) mergeLocalPlugins(c request.CTX, remoteMarketplacePlugins map[string]*model.MarketplacePlugin) *model.AppError {
	pluginsEnvironment := a.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return model.NewAppError("GetMarketplacePlugins", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError)
	}

	localPlugins, err := pluginsEnvironment.Available()
	if err != nil {
		return model.NewAppError("GetMarketplacePlugins", "app.plugin.config.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
	}

	for _, plugin := range localPlugins {
		if plugin.Manifest == nil {
			continue
		}

		if remoteMarketplacePlugins[plugin.Manifest.Id] != nil {
			// Remote plugin is installed.
			remoteMarketplacePlugins[plugin.Manifest.Id].InstalledVersion = plugin.Manifest.Version
			continue
		}

		iconData := ""
		if plugin.Manifest.IconPath != "" {
			iconData, err = getIcon(filepath.Join(plugin.Path, plugin.Manifest.IconPath))
			if err != nil {
				c.Logger().Warn("Error loading local plugin icon", mlog.String("plugin", plugin.Manifest.Id), mlog.String("icon_path", plugin.Manifest.IconPath), mlog.Err(err))
			}
		}

		var labels []model.MarketplaceLabel
		if *a.Config().PluginSettings.EnableRemoteMarketplace {
			// Labels should not (yet) be localized as the labels sent by the Marketplace are not (yet) localizable.
			labels = append(labels, model.MarketplaceLabel{
				Name:        "Local",
				Description: "This plugin is not listed in the marketplace",
			})
		}

		remoteMarketplacePlugins[plugin.Manifest.Id] = &model.MarketplacePlugin{
			BaseMarketplacePlugin: &model.BaseMarketplacePlugin{
				HomepageURL:     plugin.Manifest.HomepageURL,
				IconData:        iconData,
				ReleaseNotesURL: plugin.Manifest.ReleaseNotesURL,
				Labels:          labels,
				Manifest:        plugin.Manifest,
			},
			InstalledVersion: plugin.Manifest.Version,
		}
	}

	return nil
}

func (a *App) getBaseMarketplaceFilter(c request.CTX) *model.MarketplacePluginFilter {
	return a.ch.getBaseMarketplaceFilter(c)
}

func (ch *Channels) getBaseMarketplaceFilter(c request.CTX) *model.MarketplacePluginFilter {
	filter := &model.MarketplacePluginFilter{
		ServerVersion: model.CurrentVersion,
	}

	license := ch.srv.License(c)
	if license != nil && license.HasEnterpriseMarketplacePlugins() {
		filter.EnterprisePlugins = true
	}

	if license != nil && *license.Features.Cloud {
		filter.Cloud = true
	}

	if model.BuildEnterpriseReady == "true" {
		filter.BuildEnterpriseReady = true
	}

	filter.Platform = runtime.GOOS + "-" + runtime.GOARCH

	return filter
}

func pluginMatchesFilter(manifest *model.Manifest, filter string) bool {
	filter = strings.TrimSpace(strings.ToLower(filter))

	if filter == "" {
		return true
	}

	if strings.ToLower(manifest.Id) == filter {
		return true
	}

	if strings.Contains(strings.ToLower(manifest.Name), filter) {
		return true
	}

	if strings.Contains(strings.ToLower(manifest.Description), filter) {
		return true
	}

	return false
}

// notifyPluginEnabled notifies connected websocket clients across all peers if the version of the given
// plugin is same across them.
//
// When a peer finds itself in agreement with all other peers as to the version of the given plugin,
// it will notify all connected websocket clients (across all peers) to trigger the (re-)installation.
// There is a small chance that this never occurs, because the last server to finish installing dies before it can announce.
// There is also a chance that multiple servers notify, but the webapp handles this idempotently.
func (ch *Channels) notifyPluginEnabled(c request.CTX, manifest *model.Manifest) error {
	pluginsEnvironment := ch.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		return errors.New("pluginsEnvironment is nil")
	}
	if !manifest.HasClient() || !pluginsEnvironment.IsActive(manifest.Id) {
		return nil
	}

	var statuses model.PluginStatuses

	if ch.srv.Cluster != nil {
		var err *model.AppError
		statuses, err = ch.srv.Cluster.GetPluginStatuses()
		if err != nil {
			return err
		}
	}

	localStatus, err := ch.GetPluginStatus(manifest.Id)
	if err != nil {
		return err
	}
	statuses = append(statuses, localStatus)

	// This will not guard against the race condition of enabling a plugin immediately after installation.
	// As GetPluginStatuses() will not return the new plugin (since other peers are racing to install),
	// this peer will end up checking status against itself and will notify all webclients (including peer webclients),
	// which may result in a 404.
	for _, status := range statuses {
		if status.PluginId == manifest.Id && status.Version != manifest.Version {
			c.Logger().Debug("Not ready to notify webclients", mlog.String("cluster_id", status.ClusterId), mlog.String("plugin_id", manifest.Id))
			return nil
		}
	}

	// Notify all cluster peer clients.
	message := model.NewWebSocketEvent(model.WebsocketEventPluginEnabled, "", "", "", nil, "")
	message.Add("manifest", manifest.ClientManifest())
	ch.srv.Publish(c, message)

	return nil
}

func (ch *Channels) getPluginsFromFolder(c request.CTX) (map[string]*pluginSignaturePath, *model.AppError) {
	fileStorePaths, appErr := ch.srv.listDirectory(fileStorePluginFolder, false)
	if appErr != nil {
		return nil, model.NewAppError("getPluginsFromDir", "app.plugin.sync.list_filestore.app_error", nil, "", http.StatusInternalServerError).Wrap(appErr)
	}

	return ch.getPluginsFromFilePaths(c, fileStorePaths), nil
}

func (ch *Channels) getPluginsFromFilePaths(c request.CTX, fileStorePaths []string) map[string]*pluginSignaturePath {
	pluginSignaturePathMap := make(map[string]*pluginSignaturePath)

	fsPrefix := ""
	if *ch.cfgSvc.Config().FileSettings.DriverName == model.ImageDriverS3 {
		ptr := ch.cfgSvc.Config().FileSettings.AmazonS3PathPrefix
		if ptr != nil && *ptr != "" {
			fsPrefix = *ptr + "/"
		}
	}

	for _, path := range fileStorePaths {
		path = strings.TrimPrefix(path, fsPrefix)
		if strings.HasSuffix(path, ".tar.gz") {
			id := strings.TrimSuffix(filepath.Base(path), ".tar.gz")
			helper := &pluginSignaturePath{
				pluginID:      id,
				path:          path,
				signaturePath: "",
			}
			pluginSignaturePathMap[id] = helper
		}
	}
	for _, path := range fileStorePaths {
		path = strings.TrimPrefix(path, fsPrefix)
		if strings.HasSuffix(path, ".tar.gz.sig") {
			id := strings.TrimSuffix(filepath.Base(path), ".tar.gz.sig")
			if val, ok := pluginSignaturePathMap[id]; !ok {
				c.Logger().Warn("Unknown signature", mlog.String("path", path))
			} else {
				val.signaturePath = path
			}
		}
	}

	return pluginSignaturePathMap
}

func (ch *Channels) processPrepackagedPlugins(c request.CTX, pluginsDir string) []*plugin.PrepackagedPlugin {
	prepackagedPluginsDir, found := fileutils.FindDir(pluginsDir)
	if !found {
		return nil
	}

	var fileStorePaths []string
	err := filepath.Walk(prepackagedPluginsDir, func(walkPath string, info os.FileInfo, err error) error {
		fileStorePaths = append(fileStorePaths, walkPath)
		return nil
	})
	if err != nil {
		c.Logger().Error("Failed to walk prepackaged plugins", mlog.Err(err))
		return nil
	}

	pluginSignaturePathMap := ch.getPluginsFromFilePaths(c, fileStorePaths)
	plugins := make([]*plugin.PrepackagedPlugin, 0, len(pluginSignaturePathMap))
	prepackagedPlugins := make(chan *plugin.PrepackagedPlugin, len(pluginSignaturePathMap))

	var wg sync.WaitGroup
	for _, psPath := range pluginSignaturePathMap {
		wg.Add(1)
		go func(psPath *pluginSignaturePath) {
			defer wg.Done()
			p, err := ch.processPrepackagedPlugin(c, psPath)
			if err != nil {
				c.Logger().Error("Failed to install prepackaged plugin", mlog.String("path", psPath.path), mlog.Err(err))
				return
			}
			prepackagedPlugins <- p
		}(psPath)
	}

	wg.Wait()
	close(prepackagedPlugins)

	for p := range prepackagedPlugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// processPrepackagedPlugin will return the prepackaged plugin metadata and will also
// install the prepackaged plugin if it had been previously enabled and AutomaticPrepackagedPlugins is true.
func (ch *Channels) processPrepackagedPlugin(c request.CTX, pluginPath *pluginSignaturePath) (*plugin.PrepackagedPlugin, error) {
	c.Logger().Debug("Processing prepackaged plugin", mlog.String("path", pluginPath.path))

	fileReader, err := os.Open(pluginPath.path)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to open prepackaged plugin %s", pluginPath.path)
	}
	defer fileReader.Close()

	tmpDir, err := os.MkdirTemp("", "plugintmp")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create temp dir plugintmp")
	}
	defer os.RemoveAll(tmpDir)

	plugin, pluginDir, err := getPrepackagedPlugin(c, pluginPath, fileReader, tmpDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get prepackaged plugin %s", pluginPath.path)
	}

	// Skip installing the plugin at all if automatic prepackaged plugins is disabled
	if !*ch.cfgSvc.Config().PluginSettings.AutomaticPrepackagedPlugins {
		return plugin, nil
	}

	// Skip installing if the plugin is has not been previously enabled.
	pluginState := ch.cfgSvc.Config().PluginSettings.PluginStates[plugin.Manifest.Id]
	if pluginState == nil || !pluginState.Enable {
		return plugin, nil
	}

	c.Logger().Debug("Installing prepackaged plugin", mlog.String("path", pluginPath.path))
	if _, err := ch.installExtractedPlugin(c, plugin.Manifest, pluginDir, installPluginLocallyOnlyIfNewOrUpgrade); err != nil {
		return nil, errors.Wrapf(err, "Failed to install extracted prepackaged plugin %s", pluginPath.path)
	}

	return plugin, nil
}

// installFeatureFlagPlugins handles the automatic installation/upgrade of plugins from feature flags
func (ch *Channels) installFeatureFlagPlugins(c request.CTX) {
	ffControledPlugins := ch.cfgSvc.Config().FeatureFlags.Plugins()

	// Respect the automatic prepackaged disable setting
	if !*ch.cfgSvc.Config().PluginSettings.AutomaticPrepackagedPlugins {
		return
	}

	for pluginID, version := range ffControledPlugins {
		// Skip installing if the plugin has been previously disabled.
		pluginState := ch.cfgSvc.Config().PluginSettings.PluginStates[pluginID]
		if pluginState != nil && !pluginState.Enable {
			ch.srv.Log().Debug("Not auto installing/upgrade because plugin was disabled", mlog.String("plugin_id", pluginID), mlog.String("version", version))
			continue
		}

		// Check if we already installed this version as InstallMarketplacePlugin can't handle re-installs well.
		pluginStatus, err := ch.GetPluginStatus(pluginID)
		pluginExists := err == nil
		if pluginExists && pluginStatus.Version == version {
			continue
		}

		if version != "" && version != "control" {
			// If we are on-prem skip installation if this is a downgrade
			license := ch.srv.License(c)
			inCloud := license != nil && *license.Features.Cloud
			if !inCloud && pluginExists {
				parsedVersion, err := semver.Parse(version)
				if err != nil {
					ch.srv.Log().Debug("Bad version from feature flag", mlog.String("plugin_id", pluginID), mlog.Err(err), mlog.String("version", version))
					return
				}
				parsedExistingVersion, err := semver.Parse(pluginStatus.Version)
				if err != nil {
					ch.srv.Log().Debug("Bad version from plugin manifest", mlog.String("plugin_id", pluginID), mlog.Err(err), mlog.String("version", pluginStatus.Version))
					return
				}

				if parsedVersion.LTE(parsedExistingVersion) {
					ch.srv.Log().Debug("Skip installation because given version was a downgrade and on-prem installations should not downgrade.", mlog.String("plugin_id", pluginID), mlog.Err(err), mlog.String("version", pluginStatus.Version))
					return
				}
			}

			_, err := ch.InstallMarketplacePlugin(c, &model.InstallMarketplacePluginRequest{
				Id:      pluginID,
				Version: version,
			})
			if err != nil {
				ch.srv.Log().Debug("Unable to install plugin from FF manifest", mlog.String("plugin_id", pluginID), mlog.Err(err), mlog.String("version", version))
			} else {
				if err := ch.enablePlugin(pluginID); err != nil {
					ch.srv.Log().Debug("Unable to enable plugin installed from feature flag.", mlog.String("plugin_id", pluginID), mlog.Err(err), mlog.String("version", version))
				} else {
					ch.srv.Log().Debug("Installed and enabled plugin.", mlog.String("plugin_id", pluginID), mlog.String("version", version))
				}
			}
		}
	}
}

// getPrepackagedPlugin builds a PrepackagedPlugin from the plugin at the given path, additionally returning the directory in which it was extracted.
func getPrepackagedPlugin(c request.CTX, pluginPath *pluginSignaturePath, pluginFile io.ReadSeeker, tmpDir string) (*plugin.PrepackagedPlugin, string, error) {
	manifest, pluginDir, appErr := extractPlugin(c, pluginFile, tmpDir)
	if appErr != nil {
		return nil, "", errors.Wrapf(appErr, "Failed to extract plugin with path %s", pluginPath.path)
	}

	plugin := new(plugin.PrepackagedPlugin)
	plugin.Manifest = manifest
	plugin.Path = pluginPath.path

	if pluginPath.signaturePath != "" {
		sig := pluginPath.signaturePath
		sigReader, sigErr := os.Open(sig)
		if sigErr != nil {
			return nil, "", errors.Wrapf(sigErr, "Failed to open prepackaged plugin signature %s", sig)
		}
		bytes, sigErr := io.ReadAll(sigReader)
		if sigErr != nil {
			return nil, "", errors.Wrapf(sigErr, "Failed to read prepackaged plugin signature %s", sig)
		}
		plugin.Signature = bytes
	}

	if manifest.IconPath != "" {
		iconData, err := getIcon(filepath.Join(pluginDir, manifest.IconPath))
		if err != nil {
			c.Logger().Warn("Error loading local plugin icon", mlog.String("plugin", plugin.Manifest.Id), mlog.String("icon_path", plugin.Manifest.IconPath), mlog.Err(err))
		}
		plugin.IconData = iconData
	}

	return plugin, pluginDir, nil
}

func getIcon(iconPath string) (string, error) {
	icon, err := os.ReadFile(iconPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open icon at path %s", iconPath)
	}

	if !svg.Is(icon) {
		return "", errors.Errorf("icon is not svg %s", iconPath)
	}

	return fmt.Sprintf("data:image/svg+xml;base64,%s", base64.StdEncoding.EncodeToString(icon)), nil
}

func (ch *Channels) getPluginStateOverride(pluginID string) (bool, bool) {
	switch pluginID {
	case model.PluginIdApps:
		// Tie Apps proxy disabled status to the feature flag.
		if !ch.cfgSvc.Config().FeatureFlags.AppsEnabled {
			return true, false
		}
	case model.PluginIdCalls:
		if !ch.cfgSvc.Config().FeatureFlags.CallsEnabled {
			return true, false
		}
	}

	return false, false
}
