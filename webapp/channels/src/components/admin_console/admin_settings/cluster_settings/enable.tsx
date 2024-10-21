// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {ComponentProps} from 'react';
import React from 'react';
import {FormattedMessage} from 'react-intl';

import BooleanSetting from 'components/admin_console/boolean_setting';
import ExternalLink from 'components/external_link';

import {FIELD_IDS} from './constants';
import {messages} from './messages';

import {useIsSetByEnv} from '../hooks';

const label = <FormattedMessage {...messages.enableTitle}/>;
const helpText = (
    <FormattedMessage
        {...messages.enableDescription}
        values={{
            link: (chunks) => (
                <ExternalLink
                    href='https://docs.mattermost.com/deploy/bleve-search.html'
                    location='bleve_settings'
                >
                    {chunks}
                </ExternalLink>
            ),
        }}
    />
);

type Props = {
    value: ComponentProps<typeof BooleanSetting>['value'];
    onChange: ComponentProps<typeof BooleanSetting>['onChange'];
    isDisabled?: boolean;
}
const Enable = ({
    value,
    onChange,
    isDisabled,
}: Props) => {
    const setByEnv = useIsSetByEnv('ClusterSettings.Enable');
    return (
        <BooleanSetting
            id={FIELD_IDS.ENABLE}
            label={label}
            helpText={helpText}
            value={value}
            onChange={onChange}
            setByEnv={setByEnv}
            disabled={isDisabled}
        />
    );
};

export default Enable;