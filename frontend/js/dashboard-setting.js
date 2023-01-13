(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Sets the vue.js toggle to select and deselect panes visually.
    function settingSelectCSS(id) {
        const data = global.dashboard.$data;
        const settings = data.settings;

        for (let i = 0; i < settings.length; i++) {
            settings[i].selected = settings[i].id === id;
        }
    }


    // Selects a setting.
    global.settingSelect = function (id) {
        const data = global.dashboard.$data;
        const settings = data.settings;

        settingSelectCSS(id);

        $('ul.tabs li').removeClass('current');
        $('.content').removeClass('current');
        $('.original').addClass('current');

        for (let i = 0; i < settings.length; i++) {
            if (id === settings[i].id) {
                settings[i].open();
            }
        }
    };


    // Deselects all settings.
    global.settingDeselectAll = function () {
        const data = global.dashboard.$data;
        const settings = data.settings;

        for (let i = 0; i < settings.length; i++) {
            settings[i].selected = false;
        }
    }

}(window.commento, document));
