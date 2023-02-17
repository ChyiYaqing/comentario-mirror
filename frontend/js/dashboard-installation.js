(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Opens the installation view.
    global.installationOpen = function () {
        const html =
            `<script defer src="${global.cdn}/js/comentario.js"></script>\n` +
            '<div id="comentario"></div>\n';
        $('#code-div').text(html);
        $('pre code').each(function (i, block) {
            hljs.highlightBlock(block);
        });
        $('.view').hide();
        $('#installation-view').show();
    };

}(window.comentario, document));
