(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Opens the general settings window.
    global.generalOpen = function () {
        $('.view').hide();
        $('#general-view').show();
    };

    global.generalSaveHandler = function () {
        const data = global.dashboard.$data;

        global.buttonDisable('#save-general-button');
        global.domainUpdate(data.domains[data.cd], function () {
            global.globalOKShow('Settings saved!');
            global.buttonEnable('#save-general-button');
        });
    };

    global.ssoProviderChangeHandler = function () {
        const data = global.dashboard.$data;

        if (data.domains[data.cd].ssoSecret === '') {
            const json = {
                ownerToken: global.cookieGet('comentarioOwnerToken'),
                domain: data.domains[data.cd].domain,
            };

            global.post(`${global.origin}/api/domain/sso/new`, json, function (resp) {
                if (!resp.success) {
                    global.globalErrorShow(resp.message);
                    return;
                }

                data.domains[data.cd].ssoSecret = resp.ssoSecret;
                $('#sso-secret').val(data.domains[data.cd].ssoSecret);
            });
        } else {
            $('#sso-secret').val(data.domains[data.cd].ssoSecret);
        }
    };

}(window.comentario, document));
