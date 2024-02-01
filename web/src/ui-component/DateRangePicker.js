import PropTypes from 'prop-types';
import React from 'react';
import { Stack, Typography } from '@mui/material';
import { LocalizationProvider, DatePicker } from '@mui/x-date-pickers';
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
require('dayjs/locale/zh-cn');

export default class DateRangePicker extends React.Component {
  state = {
    startDate: this.props.defaultValue.start,
    endDate: this.props.defaultValue.end,
    localeText: this.props.localeText,
    startOpen: false,
    endOpen: false,
    views: this.props?.views
  };

  handleStartChange = (date) => {
    // 将 date设置当天的 00:00:00
    date = date.startOf('day');
    this.setState({ startDate: date });
  };

  handleEndChange = (date) => {
    // 将 date设置当天的 23:59:59
    date = date.endOf('day');
    this.setState({ endDate: date });
  };

  handleStartOpen = () => {
    this.setState({ startOpen: true });
  };

  handleStartClose = () => {
    this.setState({ startOpen: false, endOpen: true });
  };

  handleEndClose = () => {
    this.setState({ endOpen: false }, () => {
      const { startDate, endDate } = this.state;
      const { defaultValue, onChange } = this.props;
      if (!onChange) return;
      if (startDate !== defaultValue.start || endDate !== defaultValue.end) {
        onChange({ start: startDate, end: endDate });
      }
    });
  };

  render() {
    const { startOpen, endOpen, startDate, endDate, localeText } = this.state;

    return (
      <Stack direction="row" spacing={2} alignItems="center">
        <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale={'zh-cn'}>
          <DatePicker
            label={localeText?.start || ''}
            name="start_date"
            defaultValue={startDate}
            open={startOpen}
            onChange={this.handleStartChange}
            onOpen={this.handleStartOpen}
            onClose={this.handleStartClose}
            disableFuture
            disableHighlightToday
            slotProps={{
              textField: {
                readOnly: true,
                onClick: this.handleStartOpen
              }
            }}
            views={this.views}
          />
          <Typography variant="body"> – </Typography>
          <DatePicker
            label={localeText?.end || ''}
            name="end_date"
            defaultValue={endDate}
            open={endOpen}
            onChange={this.handleEndChange}
            onOpen={this.handleStartOpen}
            onClose={this.handleEndClose}
            minDate={startDate}
            disableFuture
            disableHighlightToday
            slotProps={{
              textField: {
                readOnly: true,
                onClick: this.handleStartOpen
              }
            }}
            views={this.views}
          />
        </LocalizationProvider>
      </Stack>
    );
  }
}

DateRangePicker.propTypes = {
  defaultValue: PropTypes.object,
  onChange: PropTypes.func,
  localeText: PropTypes.object,
  views: PropTypes.array
};
