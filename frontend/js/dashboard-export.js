(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    global.domainExportBegin = function () {
        const data = global.dashboard.$data;
        const json = {
            ownerToken: global.cookieGet('comentarioOwnerToken'),
            domain: data.domains[data.cd].domain,
        };

        global.buttonDisable('#domain-export-button');
        global.post(`${global.origin}/api/domain/export/begin`, json, function (resp) {
            global.buttonEnable('#domain-export-button');
            if (!resp.success) {
                global.globalErrorShow(resp.message);
                return;
            }

            global.globalOKShow('Data export operation has been successfully queued. You will receive an email.');
        });
    };

}(window.comentario, document));
