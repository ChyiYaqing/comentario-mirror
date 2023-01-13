(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Opens the import window.
    global.importOpen = function () {
        $('.view').hide();
        $('#import-view').show();
    }

    global.importDisqus = function () {
        const url = $('#disqus-url').val();
        const data = global.dashboard.$data;
        const json = {
            ownerToken: global.cookieGet('commentoOwnerToken'),
            domain: data.domains[data.cd].domain,
            url: url,
        };

        global.buttonDisable('#disqus-import-button');
        global.post(`${global.origin}/api/domain/import/disqus`, json, function (resp) {
            global.buttonEnable('#disqus-import-button');

            if (!resp.success) {
                global.globalErrorShow(resp.message);
                return;
            }

            $('#disqus-import-button').hide();

            global.globalOKShow(`Imported ${resp.numImported} comments!`);
        });
    }

    global.importCommento = function () {
        const url = $('#commento-url').val();
        const data = global.dashboard.$data;
        const json = {
            ownerToken: global.cookieGet('commentoOwnerToken'),
            domain: data.domains[data.cd].domain,
            url: url,
        };

        global.buttonDisable('#commento-import-button');
        global.post(`${global.origin}/api/domain/import/commento`, json, function (resp) {
            global.buttonEnable('#commento-import-button');

            if (!resp.success) {
                global.globalErrorShow(resp.message);
                return;
            }

            $('#commento-import-button').hide();

            global.globalOKShow(`Imported ${resp.numImported} comments!`);
        });
    }

}(window.commento, document));
