import { Wrap } from './element-wrap';
import { StringBooleanMap } from './models';
import { UIToolkit } from './ui-toolkit';
import { Dialog } from './dialog';

export class LoginDialog extends Dialog {

    private _email: Wrap<HTMLInputElement>;
    private _pwd: Wrap<HTMLInputElement>;
    private _navigateTo: string | null = null;

    constructor(
        parent: Wrap<any>,
        private readonly authMethods: StringBooleanMap,
    ) {
        super(parent);
    }

    /**
     * Entered email.
     */
    get email(): string {
        return this._email.val;
    }

    /**
     * Entered password.
     */
    get password(): string {
        return this._pwd.val;
    }

    /**
     * Where to navigate ('forgot' | 'signup') or the name of an external IdP is chosen.
     */
    get navigateTo(): string | null {
        return this._navigateTo;
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     * @param authMethods Map of enabled authentication methods.
     */
    static run(parent: Wrap<any>, authMethods: StringBooleanMap): Promise<LoginDialog> {
        const dlg = new LoginDialog(parent, authMethods);
        return dlg.run(dlg);
    }

    override renderContent(): Wrap<any> {
        // Create a login form
        const form = UIToolkit.form(() => this.dismiss(true));

        // Add OAuth buttons, if applicable
        let hasOAuth = false;
        const oauthButtons = Wrap.new('div').classes('oauth-buttons');
        const oauthProviders = ['google', 'github', 'gitlab'];
        oauthProviders.filter(p => this.authMethods[p])
            .forEach(idp => {
                Wrap.new('button')
                    .classes('button', `${idp}-button`)
                    .attr({type: 'button'})
                    .inner(idp)
                    .click(() => this.dismissWith(idp))
                    .appendTo(oauthButtons);
                hasOAuth = true;
            });

        // SSO auth
        const localAuth = this.authMethods['commento'];
        if (this.authMethods['sso']) {
            form.append(
                // SSO button
                Wrap.new('div')
                    .classes('oauth-buttons-container')
                    .append(
                        Wrap.new('div').classes('oauth-buttons')
                            .append(
                                Wrap.new('button')
                                    .classes('button', 'sso-button')
                                    .attr({type: 'button'})
                                    .inner('Single Sign-On')
                                    .click(() => this.dismissWith('sso')))),
                // Subtitle
                Wrap.new('div')
                    .classes('login-box-subtitle')
                    .inner(`Proceed with ${parent.location.host} authentication`),
                // Separator
                (hasOAuth || localAuth) && form.append(Wrap.new('hr')));
        }

        // External auth
        if (hasOAuth) {
            form.append(
                // Subtitle
                Wrap.new('div').classes('login-box-subtitle').inner('Proceed with social login'),
                // OAuth buttons
                Wrap.new('div')
                    .classes('oauth-buttons-container')
                    .append(oauthButtons),
                // Separator
                localAuth && Wrap.new('hr'));
        }

        // Local auth
        if (localAuth) {
            // Create inputs
            this._email    = UIToolkit.input('email', 'text', 'Email address', 'email');
            this._pwd = UIToolkit.input('password', 'password', 'Password', 'current-password');

            // Add the inputs to the dialog
            form.append(
                // Subtitle
                Wrap.new('div')
                    .classes('login-box-subtitle')
                    .inner('Login with your email address'),
                // Email input container
                Wrap.new('div')
                    .classes('input-container')
                    .append(Wrap.new('div').classes('input-wrapper').append(this._email)),
                // Password input container
                Wrap.new('div')
                    .classes('input-container')
                    .append(
                        Wrap.new('div')
                            .classes('input-wrapper')
                            .append(
                                this._pwd,
                                // Submit button next to the password input
                                Wrap.new('button').classes('input-button').inner('Log in').attr({type: 'submit'}))),
                // Forgot password link container
                Wrap.new('div')
                    .classes('forgot-link-container')
                    // Forgot password link
                    .append(
                        Wrap.new('a')
                            .classes('forgot-link')
                            .inner('Forgot your password?')
                            .click(() => this.dismissWith('forgot'))),
                // Switch to signup link container
                Wrap.new('div')
                    .classes('login-link-container')
                    // Switch to signup link
                    .append(
                        Wrap.new('a')
                            .classes('login-link')
                            .inner('Don\'t have an account? Sign up.')
                            .click(() => this.dismissWith('signup'))));
        }
        return form;
    }

    override onShow(): void {
        this._email.focus();
    }

    private dismissWith(nav: string) {
        this._navigateTo = nav;
        this.dismiss(true);
    }
}
