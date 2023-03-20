import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { Commenter, CommenterSettings, Email, SignupData, StringBooleanMap } from './models';
import { Utils } from './utils';
import { LoginDialog } from './login-dialog';
import { SignupDialog } from './signup-dialog';
import { SettingsDialog } from './settings-dialog';

export class ProfileBar extends Wrap<HTMLDivElement> {

    private btnSettings?: Wrap<HTMLAnchorElement>;
    private btnLogin?: Wrap<HTMLButtonElement>;
    private _authMethods?: StringBooleanMap;

    /**
     * @param baseUrl Comentario's base URL.
     * @param root Root element (for showing popups).
     * @param onLocalAuth Callback for executing a local authentication.
     * @param onOAuth Callback for executing external (OAuth) authentication.
     * @param onSignup Callback for executing user registration.
     * @param onSaveSettings Callback for saving user profile settings.
     */
    constructor(
        private readonly baseUrl: string,
        private readonly root: Wrap<any>,
        private readonly onLocalAuth: (email: string, password: string) => Promise<void>,
        private readonly onOAuth: (idp: string) => Promise<void>,
        private readonly onSignup: (data: SignupData) => Promise<void>,
        private readonly onSaveSettings: (data: CommenterSettings) => Promise<void>,
    ) {
        super(UIToolkit.div('profile-bar').element);
    }

    /**
     * Map of allowed authentication methods: {idp: true}.
     */
    set authMethods(am: StringBooleanMap | undefined) {
        this._authMethods = am;
        // Hide or show the login button based on the availability of any auth method
        this.btnLogin?.setClasses(!am || !Object.values(am).includes(true), 'hidden');
    }

    /**
     * Called whenever there's an authenticated user. Sets up the controls related to the current user.
     * @param commenter Currently authenticated user.
     * @param email Email of the commenter.
     * @param token Authenticated user's token.
     * @param onLogout Logout button click handler.
     */
    authenticated(commenter: Commenter, email: Email, token: string, onLogout: () => void): void {
        this.btnLogin = undefined;

        // Create an avatar element
        const idxColor = Utils.colourIndex(`${commenter.commenterHex}-${commenter.name}`);
        const avatar = commenter.avatarUrl ?
            Wrap.new('img')
                .classes('avatar-img')
                .attr({
                    src: `${this.baseUrl}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                    loading: 'lazy',
                    alt: '',
                }) :
            UIToolkit.div('avatar', `bg-${idxColor}`).html(commenter.name![0].toUpperCase());

        // Recreate the content
        this.html('')
            .append(
                // Commenter avatar and name
                UIToolkit.div('logged-in-as')
                    .append(
                        // Avatar
                        avatar,
                        // Name and link
                        Wrap.new(commenter.websiteUrl ? 'a' : 'div')
                            .classes('name')
                            .inner(commenter.name!)
                            .attr({
                                href: commenter.websiteUrl,
                                rel:  commenter.websiteUrl && 'nofollow noopener noreferrer',
                            })),
                // Buttons on the right
                UIToolkit.div()
                    .append(
                        // Settings link
                        this.btnSettings = Wrap.new('a')
                            .classes('profile-link')
                            .inner('Settings')
                            .click((_, e) => {
                                // Prevent the page from being reloaded because of the empty href
                                e.preventDefault();
                                return this.editSettings(commenter);
                            }),
                        // Logout link
                        Wrap.new('a')
                            .classes('profile-link')
                            .inner('Logout')
                            .attr({href: ''})
                            .click((_, e) => {
                                // Prevent the page from being reloaded because of the empty href
                                e.preventDefault();
                                onLogout();
                            })));
    }

    /**
     * Called whenever there's no authenticated user. Sets up the login controls.
     */
    notAuthenticated(): void {
        // Remove all content
        this.html('')
            .append(
                // Add an empty div to push the button to the right (profile bar uses 'justify-content: space-between')
                UIToolkit.div(),
                // Add a Login button
                this.btnLogin = UIToolkit.button('Login', () => this.loginUser(), 'fw-bold'));
    }

    /**
     * Show a login dialog and return a promise that's resolved when the dialog is closed.
     */
    async loginUser(): Promise<void> {
        if (!this._authMethods) {
            return Promise.reject('No configured authentication methods.');
        }
        const dlg = await LoginDialog.run(
            this.root,
            {ref: this.btnLogin!, placement: 'bottom-end'},
            this._authMethods,
            this.baseUrl);
        if (dlg.confirmed) {
            switch (dlg.navigateTo) {
                case null:
                    // Local auth
                    return this.onLocalAuth(dlg.email, dlg.password);

                case 'forgot':
                    // Already navigated to the Forgot password page in a new tab
                    return;

                case 'signup':
                    // Switch to signup
                    return this.signupUser();

                default:
                    // External auth
                    return this.onOAuth(dlg.navigateTo);
            }
        }
    }

    /**
     * Show a signup dialog and return a promise that's resolved when the dialog is closed.
     */
    async signupUser(): Promise<void> {
        const dlg = await SignupDialog.run(this.root, {ref: this.btnLogin!, placement: 'bottom-end'});
        if (dlg.confirmed) {
            await this.onSignup(dlg.data);
        }
    }

    /**
     * Show the settings dialog and return a promise that's resolved when the dialog is closed.
     */
    async editSettings(commenter: Commenter): Promise<void> {
        const dlg = await SettingsDialog.run(this.root, {ref: this.btnSettings!, placement: 'bottom-end'}, commenter);
        if (dlg.confirmed) {
            await this.onSaveSettings(dlg.data);
        }
    }
}
