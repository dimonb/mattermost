// Code generated by mockery v2.42.2. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "github.com/mattermost/mattermost/server/public/model"
	mock "github.com/stretchr/testify/mock"
)

// WebhookStore is an autogenerated mock type for the WebhookStore type
type WebhookStore struct {
	mock.Mock
}

// AnalyticsIncomingCount provides a mock function with given fields: teamID, userID
func (_m *WebhookStore) AnalyticsIncomingCount(teamID string, userID string) (int64, error) {
	ret := _m.Called(teamID, userID)

	if len(ret) == 0 {
		panic("no return value specified for AnalyticsIncomingCount")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (int64, error)); ok {
		return rf(teamID, userID)
	}
	if rf, ok := ret.Get(0).(func(string, string) int64); ok {
		r0 = rf(teamID, userID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(teamID, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AnalyticsOutgoingCount provides a mock function with given fields: teamID
func (_m *WebhookStore) AnalyticsOutgoingCount(teamID string) (int64, error) {
	ret := _m.Called(teamID)

	if len(ret) == 0 {
		panic("no return value specified for AnalyticsOutgoingCount")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (int64, error)); ok {
		return rf(teamID)
	}
	if rf, ok := ret.Get(0).(func(string) int64); ok {
		r0 = rf(teamID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(teamID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ClearCaches provides a mock function with given fields:
func (_m *WebhookStore) ClearCaches() {
	_m.Called()
}

// DeleteIncoming provides a mock function with given fields: webhookID, timestamp
func (_m *WebhookStore) DeleteIncoming(webhookID string, timestamp int64) error {
	ret := _m.Called(webhookID, timestamp)

	if len(ret) == 0 {
		panic("no return value specified for DeleteIncoming")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(webhookID, timestamp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteOutgoing provides a mock function with given fields: webhookID, timestamp
func (_m *WebhookStore) DeleteOutgoing(webhookID string, timestamp int64) error {
	ret := _m.Called(webhookID, timestamp)

	if len(ret) == 0 {
		panic("no return value specified for DeleteOutgoing")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int64) error); ok {
		r0 = rf(webhookID, timestamp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetIncoming provides a mock function with given fields: id, allowFromCache
func (_m *WebhookStore) GetIncoming(id string, allowFromCache bool) (*model.IncomingWebhook, error) {
	ret := _m.Called(id, allowFromCache)

	if len(ret) == 0 {
		panic("no return value specified for GetIncoming")
	}

	var r0 *model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, bool) (*model.IncomingWebhook, error)); ok {
		return rf(id, allowFromCache)
	}
	if rf, ok := ret.Get(0).(func(string, bool) *model.IncomingWebhook); ok {
		r0 = rf(id, allowFromCache)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, bool) error); ok {
		r1 = rf(id, allowFromCache)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIncomingByChannel provides a mock function with given fields: channelID
func (_m *WebhookStore) GetIncomingByChannel(channelID string) ([]*model.IncomingWebhook, error) {
	ret := _m.Called(channelID)

	if len(ret) == 0 {
		panic("no return value specified for GetIncomingByChannel")
	}

	var r0 []*model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]*model.IncomingWebhook, error)); ok {
		return rf(channelID)
	}
	if rf, ok := ret.Get(0).(func(string) []*model.IncomingWebhook); ok {
		r0 = rf(channelID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(channelID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIncomingByTeam provides a mock function with given fields: teamID, offset, limit
func (_m *WebhookStore) GetIncomingByTeam(teamID string, offset int, limit int) ([]*model.IncomingWebhook, error) {
	ret := _m.Called(teamID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetIncomingByTeam")
	}

	var r0 []*model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, int, int) ([]*model.IncomingWebhook, error)); ok {
		return rf(teamID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.IncomingWebhook); ok {
		r0 = rf(teamID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(teamID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIncomingByTeamByUser provides a mock function with given fields: teamID, userID, offset, limit
func (_m *WebhookStore) GetIncomingByTeamByUser(teamID string, userID string, offset int, limit int) ([]*model.IncomingWebhook, error) {
	ret := _m.Called(teamID, userID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetIncomingByTeamByUser")
	}

	var r0 []*model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, int, int) ([]*model.IncomingWebhook, error)); ok {
		return rf(teamID, userID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, string, int, int) []*model.IncomingWebhook); ok {
		r0 = rf(teamID, userID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, int, int) error); ok {
		r1 = rf(teamID, userID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIncomingList provides a mock function with given fields: offset, limit
func (_m *WebhookStore) GetIncomingList(offset int, limit int) ([]*model.IncomingWebhook, error) {
	ret := _m.Called(offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetIncomingList")
	}

	var r0 []*model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(int, int) ([]*model.IncomingWebhook, error)); ok {
		return rf(offset, limit)
	}
	if rf, ok := ret.Get(0).(func(int, int) []*model.IncomingWebhook); ok {
		r0 = rf(offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIncomingListByUser provides a mock function with given fields: userID, offset, limit
func (_m *WebhookStore) GetIncomingListByUser(userID string, offset int, limit int) ([]*model.IncomingWebhook, error) {
	ret := _m.Called(userID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetIncomingListByUser")
	}

	var r0 []*model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, int, int) ([]*model.IncomingWebhook, error)); ok {
		return rf(userID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.IncomingWebhook); ok {
		r0 = rf(userID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(userID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoing provides a mock function with given fields: id
func (_m *WebhookStore) GetOutgoing(id string) (*model.OutgoingWebhook, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoing")
	}

	var r0 *model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*model.OutgoingWebhook, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) *model.OutgoingWebhook); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoingByChannel provides a mock function with given fields: channelID, offset, limit
func (_m *WebhookStore) GetOutgoingByChannel(channelID string, offset int, limit int) ([]*model.OutgoingWebhook, error) {
	ret := _m.Called(channelID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoingByChannel")
	}

	var r0 []*model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, int, int) ([]*model.OutgoingWebhook, error)); ok {
		return rf(channelID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.OutgoingWebhook); ok {
		r0 = rf(channelID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(channelID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoingByChannelByUser provides a mock function with given fields: channelID, userID, offset, limit
func (_m *WebhookStore) GetOutgoingByChannelByUser(channelID string, userID string, offset int, limit int) ([]*model.OutgoingWebhook, error) {
	ret := _m.Called(channelID, userID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoingByChannelByUser")
	}

	var r0 []*model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, int, int) ([]*model.OutgoingWebhook, error)); ok {
		return rf(channelID, userID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, string, int, int) []*model.OutgoingWebhook); ok {
		r0 = rf(channelID, userID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, int, int) error); ok {
		r1 = rf(channelID, userID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoingByTeam provides a mock function with given fields: teamID, offset, limit
func (_m *WebhookStore) GetOutgoingByTeam(teamID string, offset int, limit int) ([]*model.OutgoingWebhook, error) {
	ret := _m.Called(teamID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoingByTeam")
	}

	var r0 []*model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, int, int) ([]*model.OutgoingWebhook, error)); ok {
		return rf(teamID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.OutgoingWebhook); ok {
		r0 = rf(teamID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(teamID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoingByTeamByUser provides a mock function with given fields: teamID, userID, offset, limit
func (_m *WebhookStore) GetOutgoingByTeamByUser(teamID string, userID string, offset int, limit int) ([]*model.OutgoingWebhook, error) {
	ret := _m.Called(teamID, userID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoingByTeamByUser")
	}

	var r0 []*model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, int, int) ([]*model.OutgoingWebhook, error)); ok {
		return rf(teamID, userID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, string, int, int) []*model.OutgoingWebhook); ok {
		r0 = rf(teamID, userID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, int, int) error); ok {
		r1 = rf(teamID, userID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoingList provides a mock function with given fields: offset, limit
func (_m *WebhookStore) GetOutgoingList(offset int, limit int) ([]*model.OutgoingWebhook, error) {
	ret := _m.Called(offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoingList")
	}

	var r0 []*model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(int, int) ([]*model.OutgoingWebhook, error)); ok {
		return rf(offset, limit)
	}
	if rf, ok := ret.Get(0).(func(int, int) []*model.OutgoingWebhook); ok {
		r0 = rf(offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(int, int) error); ok {
		r1 = rf(offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOutgoingListByUser provides a mock function with given fields: userID, offset, limit
func (_m *WebhookStore) GetOutgoingListByUser(userID string, offset int, limit int) ([]*model.OutgoingWebhook, error) {
	ret := _m.Called(userID, offset, limit)

	if len(ret) == 0 {
		panic("no return value specified for GetOutgoingListByUser")
	}

	var r0 []*model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(string, int, int) ([]*model.OutgoingWebhook, error)); ok {
		return rf(userID, offset, limit)
	}
	if rf, ok := ret.Get(0).(func(string, int, int) []*model.OutgoingWebhook); ok {
		r0 = rf(userID, offset, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(string, int, int) error); ok {
		r1 = rf(userID, offset, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InvalidateWebhookCache provides a mock function with given fields: webhook
func (_m *WebhookStore) InvalidateWebhookCache(webhook string) {
	_m.Called(webhook)
}

// MergeIncomingWebhookUserId provides a mock function with given fields: toUserID, fromUserID
func (_m *WebhookStore) MergeIncomingWebhookUserId(toUserID string, fromUserID string) error {
	ret := _m.Called(toUserID, fromUserID)

	if len(ret) == 0 {
		panic("no return value specified for MergeIncomingWebhookUserId")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(toUserID, fromUserID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MergeOutgoingWebhookUserId provides a mock function with given fields: toUserID, fromUserID
func (_m *WebhookStore) MergeOutgoingWebhookUserId(toUserID string, fromUserID string) error {
	ret := _m.Called(toUserID, fromUserID)

	if len(ret) == 0 {
		panic("no return value specified for MergeOutgoingWebhookUserId")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(toUserID, fromUserID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PermanentDeleteIncomingByChannel provides a mock function with given fields: channelID
func (_m *WebhookStore) PermanentDeleteIncomingByChannel(channelID string) error {
	ret := _m.Called(channelID)

	if len(ret) == 0 {
		panic("no return value specified for PermanentDeleteIncomingByChannel")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(channelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PermanentDeleteIncomingByUser provides a mock function with given fields: userID
func (_m *WebhookStore) PermanentDeleteIncomingByUser(userID string) error {
	ret := _m.Called(userID)

	if len(ret) == 0 {
		panic("no return value specified for PermanentDeleteIncomingByUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PermanentDeleteOutgoingByChannel provides a mock function with given fields: channelID
func (_m *WebhookStore) PermanentDeleteOutgoingByChannel(channelID string) error {
	ret := _m.Called(channelID)

	if len(ret) == 0 {
		panic("no return value specified for PermanentDeleteOutgoingByChannel")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(channelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PermanentDeleteOutgoingByUser provides a mock function with given fields: userID
func (_m *WebhookStore) PermanentDeleteOutgoingByUser(userID string) error {
	ret := _m.Called(userID)

	if len(ret) == 0 {
		panic("no return value specified for PermanentDeleteOutgoingByUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveIncoming provides a mock function with given fields: webhook
func (_m *WebhookStore) SaveIncoming(webhook *model.IncomingWebhook) (*model.IncomingWebhook, error) {
	ret := _m.Called(webhook)

	if len(ret) == 0 {
		panic("no return value specified for SaveIncoming")
	}

	var r0 *model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.IncomingWebhook) (*model.IncomingWebhook, error)); ok {
		return rf(webhook)
	}
	if rf, ok := ret.Get(0).(func(*model.IncomingWebhook) *model.IncomingWebhook); ok {
		r0 = rf(webhook)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.IncomingWebhook) error); ok {
		r1 = rf(webhook)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveOutgoing provides a mock function with given fields: webhook
func (_m *WebhookStore) SaveOutgoing(webhook *model.OutgoingWebhook) (*model.OutgoingWebhook, error) {
	ret := _m.Called(webhook)

	if len(ret) == 0 {
		panic("no return value specified for SaveOutgoing")
	}

	var r0 *model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.OutgoingWebhook) (*model.OutgoingWebhook, error)); ok {
		return rf(webhook)
	}
	if rf, ok := ret.Get(0).(func(*model.OutgoingWebhook) *model.OutgoingWebhook); ok {
		r0 = rf(webhook)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.OutgoingWebhook) error); ok {
		r1 = rf(webhook)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateIncoming provides a mock function with given fields: webhook
func (_m *WebhookStore) UpdateIncoming(webhook *model.IncomingWebhook) (*model.IncomingWebhook, error) {
	ret := _m.Called(webhook)

	if len(ret) == 0 {
		panic("no return value specified for UpdateIncoming")
	}

	var r0 *model.IncomingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.IncomingWebhook) (*model.IncomingWebhook, error)); ok {
		return rf(webhook)
	}
	if rf, ok := ret.Get(0).(func(*model.IncomingWebhook) *model.IncomingWebhook); ok {
		r0 = rf(webhook)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.IncomingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.IncomingWebhook) error); ok {
		r1 = rf(webhook)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateOutgoing provides a mock function with given fields: hook
func (_m *WebhookStore) UpdateOutgoing(hook *model.OutgoingWebhook) (*model.OutgoingWebhook, error) {
	ret := _m.Called(hook)

	if len(ret) == 0 {
		panic("no return value specified for UpdateOutgoing")
	}

	var r0 *model.OutgoingWebhook
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.OutgoingWebhook) (*model.OutgoingWebhook, error)); ok {
		return rf(hook)
	}
	if rf, ok := ret.Get(0).(func(*model.OutgoingWebhook) *model.OutgoingWebhook); ok {
		r0 = rf(hook)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.OutgoingWebhook)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.OutgoingWebhook) error); ok {
		r1 = rf(hook)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWebhookStore creates a new instance of WebhookStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWebhookStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *WebhookStore {
	mock := &WebhookStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
