(function (global, document) {
    'use strict';

    global.logout = function () {
        global.cookieDelete('comentarioOwnerToken');
        document.location = `${global.origin}/login`;
    };

}(window.comentario, document));
