export function getLastSevenDays() {
  const dates = [];
  for (let i = 6; i >= 0; i--) {
    const d = new Date();
    d.setDate(d.getDate() - i);
    const month = '' + (d.getMonth() + 1);
    const day = '' + d.getDate();
    const year = d.getFullYear();

    const formattedDate = [year, month.padStart(2, '0'), day.padStart(2, '0')].join('-');
    dates.push(formattedDate);
  }
  return dates;
}

export function getTodayDay() {
  let today = new Date();
  return today.toISOString().slice(0, 10);
}

export function generateLineChartOptions(data, unit) {
  const dates = data.map((item) => item.date);
  const values = data.map((item) => item.value);

  const minDate = dates[0];
  const maxDate = dates[dates.length - 1];

  const minValue = Math.min(...values);
  const maxValue = Math.max(...values);

  return {
    series: [
      {
        data: values
      }
    ],
    type: 'line',
    height: 90,
    options: {
      chart: {
        sparkline: {
          enabled: true
        }
      },
      dataLabels: {
        enabled: false
      },
      colors: ['#fff'],
      fill: {
        type: 'solid',
        opacity: 1
      },
      stroke: {
        curve: 'smooth',
        width: 3
      },
      xaxis: {
        categories: dates,
        labels: {
          show: false
        },
        min: minDate,
        max: maxDate
      },
      yaxis: {
        min: minValue,
        max: maxValue,
        labels: {
          show: false
        }
      },
      tooltip: {
        theme: 'dark',
        fixed: {
          enabled: false
        },
        x: {
          format: 'yyyy-MM-dd'
        },
        y: {
          formatter: function (val) {
            return val + ` ${unit}`;
          },
          title: {
            formatter: function () {
              return '';
            }
          }
        },
        marker: {
          show: false
        }
      }
    }
  };
}

export function generateBarChartOptions(xaxis, data, unit = '', decimal = 0) {
  return {
    height: 480,
    type: 'bar',
    options: {
      title: {
        align: 'left',
        style: {
          fontSize: '14px',
          fontWeight: 'bold',
          fontFamily: 'Roboto, sans-serif'
        }
      },
      colors: [
        '#008FFB',
        '#00E396',
        '#FEB019',
        '#FF4560',
        '#775DD0',
        '#55efc4',
        '#81ecec',
        '#74b9ff',
        '#a29bfe',
        '#00b894',
        '#00cec9',
        '#0984e3',
        '#6c5ce7',
        '#ffeaa7',
        '#fab1a0',
        '#ff7675',
        '#fd79a8',
        '#fdcb6e',
        '#e17055',
        '#d63031',
        '#e84393'
      ],
      chart: {
        id: 'bar-chart',
        stacked: true,
        toolbar: {
          show: true
        },
        zoom: {
          enabled: true
        }
      },
      responsive: [
        {
          breakpoint: 480,
          options: {
            legend: {
              position: 'bottom',
              offsetX: -10,
              offsetY: 0
            }
          }
        }
      ],
      plotOptions: {
        bar: {
          horizontal: false,
          columnWidth: '50%',
          // borderRadius: 10,
          dataLabels: {
            total: {
              enabled: true,
              style: {
                fontSize: '13px',
                fontWeight: 900
              },
              formatter: function (val) {
                return renderChartNumber(val, decimal);
              }
            }
          }
        }
      },
      xaxis: {
        type: 'category',
        categories: xaxis
      },
      legend: {
        show: true,
        fontSize: '14px',
        fontFamily: `'Roboto', sans-serif`,
        position: 'bottom',
        offsetX: 20,
        labels: {
          useSeriesColors: false
        },
        markers: {
          width: 16,
          height: 16,
          radius: 5
        },
        itemMargin: {
          horizontal: 15,
          vertical: 8
        }
      },
      fill: {
        type: 'solid'
      },
      dataLabels: {
        enabled: false
      },
      grid: {
        show: true
      },
      tooltip: {
        theme: 'dark',
        fixed: {
          enabled: false
        },
        marker: {
          show: false
        },
        y: {
          formatter: function (val) {
            return renderChartNumber(val, decimal) + ` ${unit}`;
          }
        }
      }
    },
    series: data
  };
}

// 格式化数值
export function renderChartNumber(number, decimal = 2) {
  number = number.toFixed(decimal);
  if (number === Number(0).toFixed(decimal)) {
    return 0;
  }

  // 如果大于1000，显示为k
  if (number >= 1000) {
    return (number / 1000).toFixed(1) + 'k';
  }

  return number;
}
