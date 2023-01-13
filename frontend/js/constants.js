(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    if (window.commento === undefined) {
        window.commento = {};
    }

    window.commento.origin = '[[[.Origin]]]';
    window.commento.cdn = '[[[.CdnPrefix]]]';

}(window, document));
