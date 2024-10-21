// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {ComponentProps} from 'react';
import React from 'react';
import {defineMessage, FormattedMessage} from 'react-intl';

import TextSetting from 'components/admin_console/text_setting';

import {FIELD_IDS} from './constants';
import {messages} from './messages';

import {useIsSetByEnv} from '../hooks';

const label = <FormattedMessage {...messages.maxConnectionsTitle}/>;
const helpText = <FormattedMessage {...messages.maxConnectionsDescription}/>;
const placeholder = defineMessage({
    id: 'admin.sql.maxConnectionsExample',
    defaultMessage: 'E.g.: "10"',
});

type Props = {
    value: ComponentProps<typeof TextSetting>['value'];
    onChange: ComponentProps<typeof TextSetting>['onChange'];
    isDisabled?: ComponentProps<typeof TextSetting>['disabled'];
}

const MaxIdleConns = ({
    onChange,
    value,
    isDisabled,
}: Props) => {
    const setByEnv = useIsSetByEnv('SqlSettings.MaxIdleConns');
    return (
        <TextSetting
            id={FIELD_IDS.MAX_IDLE_CONNS}
            placeholder={placeholder}
            label={label}
            helpText={helpText}
            value={value}
            onChange={onChange}
            setByEnv={setByEnv}
            disabled={isDisabled}
            type='text'
        />
    );
};

export default MaxIdleConns;