(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Performs a JSON POST request to the given url with the given data and
    // calls the callback function with the JSON response.
    global.post = function (url, json, callback) {
        $.ajax({
            url: url,
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(json),
            success: callback,
        });
    };

    // Performs a GET request and calls the callback function with the JSON
    // response.
    global.get = function (url, callback) {
        $.ajax({
            url: url,
            type: 'GET',
            success: function (data) {
                callback(JSON.parse(data));
            },
        });
    };
}(window.commento, document));
