(function (global, document) {
    'use strict';

    global.resetPassword = function (event) {
        event.preventDefault();

        const allOk = global.unfilledMark(['#password', '#password2'], function (el) {
            el.css('border-bottom', '1px solid red');
        });

        if (!allOk) {
            global.textSet('#err', 'Please make sure all fields are filled.');
            return;
        }

        const pwd = $('#password');
        if (pwd.val() !== $('#password2').val()) {
            global.textSet('#err', 'The two passwords do not match.');
            return;
        }

        const json = {
            'resetHex': global.paramGet('hex'),
            'password': pwd.val(),
        };

        global.buttonDisable('#reset-button');
        global.post(global.origin + '/api/reset', json, function (resp) {
            global.buttonEnable('#reset-button');

            global.textSet('#err', '');
            if (!resp.success) {
                global.textSet('#err', resp.message);
                return
            }

            if (resp.entity === 'owner') {
                document.location = global.origin + '/login?changed=true';
            } else {
                $('#msg').html('Your password has been reset. You may close this window and try logging in again.');
            }
        });
    }

}(window.commento, document));
