(function (global, document) {
    'use strict';

    // TODO No-op statement to prevent the IDE from complaining about unused function argument
    // noinspection BadExpressionStatementJS
    (document);

    // Shows messages produced from email confirmation attempts.
    function displayConfirmedEmail() {
        const confirmed = global.paramGet('confirmed');

        if (confirmed === 'true') {
            $('#msg').html('Successfully confirmed! Login to continue.')
        } else if (confirmed === 'false') {
            $('#err').html('That link has expired.')
        }
    }


    // Shows messages produced from password reset attempts.
    function displayChangedPassword() {
        const changed = global.paramGet('changed');

        if (changed === 'true') {
            $('#msg').html('Password changed successfully! Login to continue.')
        }
    }

    // Shows messages produced from completed signups.
    function displaySignedUp() {
        const signedUp = global.paramGet('signedUp');

        if (signedUp === 'true') {
            $('#msg').html('Registration successful! Login to continue.')
        }
    }

    // Shows messages produced from account deletion.
    function displayDeletedOwner() {
        const deleted = global.paramGet('deleted');

        if (deleted === 'true') {
            $('#msg').html('Your account has been deleted.')
        }
    }

    // Shows email confirmation and password reset messages, if any.
    global.displayMessages = function () {
        displayConfirmedEmail();
        displayChangedPassword();
        displaySignedUp();
        displayDeletedOwner();
    };

    // Logs the user in and redirects to the dashboard.
    global.login = function (event) {
        event.preventDefault();

        const allOk = global.unfilledMark(['#email', '#password'], function (el) {
            el.css('border-bottom', '1px solid red');
        });

        if (!allOk) {
            global.textSet('#err', 'Please make sure all fields are filled');
            return;
        }

        const json = {
            email: $('#email').val(),
            password: $('#password').val(),
        };

        global.buttonDisable('#login-button');
        global.post(`${global.origin}/api/owner/login`, json, function (resp) {
            global.buttonEnable('#login-button');

            if (!resp.success) {
                global.textSet('#err', resp.message);
                return;
            }

            global.cookieSet('commentoOwnerToken', resp.ownerToken);
            document.location = `${global.origin}/dashboard`;
        });
    };

}(window.commento, document));
