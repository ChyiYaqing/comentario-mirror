import { Wrap } from './element-wrap';
import { StringBooleanMap } from './models';

export enum LoginStatus {
    /** User dismissed the dialog. */
    CANCEL,
    /** User confirmed (submitted) the dialog. */
    OK,
    /** User clicked on the "Forgot Password" link. */
    GO_TO_FORGOT_PASSWORD,
    /** User clicked on the "Sign up" link. */
    GO_TO_SIGNUP,
}

/**
 * The outcome of the login dialog.
 */
export interface LoginResult {
    /** Status of the dialog. */
    readonly status: LoginStatus;
    /** External identity provider, if chosen. */
    readonly idp?: string;
    /** Email (local auth). */
    readonly email?: string;
    /** Password (local auth). */
    readonly password?: string;
}

export class LoginDialog {

    private backdrop: Wrap<HTMLDivElement>;
    private container: Wrap<HTMLDivElement>;
    private dialogBox: Wrap<HTMLFormElement>;
    private emailInput: Wrap<HTMLInputElement>;
    private passwordInput: Wrap<HTMLInputElement>;

    constructor(
        private readonly parent: Wrap<any>,
        private readonly authMethods: StringBooleanMap,
        private readonly resolve: (r: LoginResult | PromiseLike<LoginResult>) => void,
    ) {}

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed, with the outcome
     * of the operation.
     * @param parent Parent element for the dialog.
     * @param authMethods Map of enabled authentication methods.
     */
    static run(parent: Wrap<any>, authMethods: StringBooleanMap): Promise<LoginResult> {
        return new Promise(resolve => new LoginDialog(parent, authMethods, resolve).render());
    }

    private render() {
        // Create a login box
        this.dialogBox = Wrap.new('form')
            .classes('login-box', 'fade-in')
            // Form submit event
            .on('submit', e => {
                e.preventDefault();
                this.dismiss(LoginStatus.OK);
            })
            // Don't propagate the click to prevent cancelling the dialog, which happens when the click reaches the
            // parent container
            .click(e => e.stopPropagation())
            // Close button
            .append(
                Wrap.new('button')
                    .classes('btn-login-box-close')
                    .attr({type: 'button', ariaLabel: 'Close'})
                    .click(() => this.dismiss(LoginStatus.CANCEL)));

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
                    .click(() => this.dismiss(LoginStatus.OK, idp))
                    .appendTo(oauthButtons);
                hasOAuth = true;
            });

        // SSO auth
        const localAuth = this.authMethods['commento'];
        if (this.authMethods['sso']) {
            this.dialogBox.append(
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
                                    .click(() => this.dismiss(LoginStatus.OK, 'sso')))),
                // Subtitle
                Wrap.new('div')
                    .classes('login-box-subtitle')
                    .inner(`Proceed with ${parent.location.host} authentication`),
                // Separator
                (hasOAuth || localAuth) && this.dialogBox.append(Wrap.new('hr')));
        }

        // External auth
        if (hasOAuth) {
            this.dialogBox.append(
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
            this.emailInput = Wrap.new('input')
                .classes('input')
                .attr({name: 'email', placeholder: 'Email address', type: 'text', autocomplete: 'email'});
            this.passwordInput = Wrap.new('input')
                .classes('input')
                .attr({name: 'password', type: 'password', placeholder: 'Password', autocomplete: 'current-password'});

            // Add the inputs to the dialog
            this.dialogBox.append(
                // Subtitle
                Wrap.new('div')
                    .classes('login-box-subtitle')
                    .inner('Login with your email address'),
                // Email input container
                Wrap.new('div')
                    .classes('email-container')
                    .append(Wrap.new('div').classes('email').append(this.emailInput)),
                // Password input container
                Wrap.new('div')
                    .classes('email-container')
                    .append(
                        Wrap.new('div')
                            .classes('email')
                            .append(
                                this.passwordInput,
                                // Submit button next to the password input
                                Wrap.new('button').classes('email-button').inner('Log in').attr({type: 'submit'}))),
                // Forgot password link container
                Wrap.new('div')
                    .classes('forgot-link-container')
                    // Forgot password link
                    .append(
                        Wrap.new('a')
                            .classes('forgot-link')
                            .inner('Forgot your password?')
                            .click(() => this.dismiss(LoginStatus.GO_TO_FORGOT_PASSWORD))),
                // Switch to signup link container
                Wrap.new('div')
                    .classes('login-link-container')
                    // Switch to signup link
                    .append(
                        Wrap.new('a')
                            .classes('login-link')
                            .inner('Don\'t have an account? Sign up.')
                            .click(() => this.dismiss(LoginStatus.GO_TO_SIGNUP))));
        }

        // Create a backdrop
        this.backdrop = Wrap.new('div').classes('backdrop').appendTo(this.parent);

        // Add the dialog to the container and scroll to it, if necessary
        this.container = Wrap.new('div')
            .classes('login-box-container')
            // Cancel the dialog when clicked outside
            .click(() => this.dismiss(LoginStatus.CANCEL))
            .appendTo(this.parent);
        this.dialogBox.appendTo(this.container).scrollTo();

        // Focus the email input
        this.emailInput.focus();
    }

    private dismiss(status: LoginStatus, idp?: string) {
        const email = this.emailInput.val;
        const password = this.passwordInput.val;

        // Close the dialog
        this.dialogBox.noClasses('fade-in').classes('fade-out');
        setTimeout(
            () => {
                this.container.remove();
                this.backdrop.remove();

                // Resolve the promise, returning the result
                this.resolve({status, idp, email, password});
            },
            250);
    }
}
