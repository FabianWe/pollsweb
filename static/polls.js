/**
 Copyright 2020 Fabian Wenzelmann <fabianwen@posteo.eu>

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
 * */

// global setup
moment.tz.setDefault(timeZone);

$.fn.datetimepicker.Constructor.Default = $.extend({}, $.fn.datetimepicker.Constructor.Default, {
    icons: {
        time: 'fas fa-clock',
        date: 'fas fa-calendar',
        up: 'fas fa-arrow-up',
        down: 'fas fa-arrow-down',
        previous: 'fas fa-chevron-left',
        next: 'fas fa-chevron-right',
        today: 'fas fa-calendar-check',
        clear: 'fas fa-trash',
        close: 'fas fa-times'
    },
    format: momentDateTimeFormat,
    timeZone: timeZone
});

function formatMomentDatetimeTransfer(m) {
    return m.utc().format(momentTransferDatetimeFormat);
}

// TODO bind an event to avoid end < start?
function initPeriodForm(formPrefix) {
    let startName = '#' + formPrefix + 'Start';
    let endName = '#' + formPrefix + 'End';
    let timeName = '#' + formPrefix + 'Time';
    $(startName).datetimepicker({
            stepping: 5
        }
    );
    $(endName).datetimepicker({
            stepping: 5
        }
    );
    $(timeName).datetimepicker({
        stepping: 5,
        format: 'HH:mm'
    });
    $('#'+ formPrefix).submit(function (e) {
        e.preventDefault();
        console.log("Here we are!");
        let form = $('#' + formPrefix);
        let data = form.serializeArray();
        let fieldMap = new Map();
        data.forEach(function (value) {
            fieldMap.set(value.name, value.value);
        });
        let start = $(startName);
        let end = $(endName);
        let startString = formatMomentDatetimeTransfer(start.datetimepicker('viewDate'));
        let endString = formatMomentDatetimeTransfer(end.datetimepicker('viewDate'));
        let queryData = {
            period_name: fieldMap.get('period_name'),
            period_start: startString,
            period_end: endString,
            weekday: fieldMap.get('weekday'),
            time: fieldMap.get('time')
        };
        console.log('?' + $.param(queryData) );
        $.post('?' + $.param(queryData) );
    });
}
