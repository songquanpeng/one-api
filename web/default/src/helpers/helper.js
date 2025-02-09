import {CHANNEL_OPTIONS} from '../constants';

let channelMap = undefined;

export function getChannelOption(channelId) {
    if (channelMap === undefined) {
        channelMap = {};
        CHANNEL_OPTIONS.forEach((option) => {
            channelMap[option.key] = option;
        });
    }
    return channelMap[channelId];
}
