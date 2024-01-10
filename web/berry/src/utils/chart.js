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

export function generateChartOptions(data, unit) {
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
