(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    if (window.comentario === undefined) {
        window.comentario = {};
    }

    window.comentario.origin = '[[[.Origin]]]';
    window.comentario.cdn = '[[[.CdnPrefix]]]';

}(window, document));
