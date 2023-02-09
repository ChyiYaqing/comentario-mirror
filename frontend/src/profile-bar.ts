import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { Commenter, Email, StringBooleanMap } from './models';
import { Utils } from './utils';
import { LoginDialog } from './login-dialog';
import { SignupDialog } from './signup-dialog';

export class ProfileBar extends Wrap<HTMLDivElement> {

    private btnLogin?: Wrap<HTMLButtonElement>;
    private _authMethods?: StringBooleanMap;

    /**
     * @param cdn Comentario's CDN URL.
     * @param origin Comentario's origin URL.
     * @param root Root element (for showing popups).
     * @param onLocalAuth Callback for executing a local authentication.
     * @param onOAuth Callback for executing external (OAuth) authentication.
     * @param onSignup Callback for executing user registration.
     */
    constructor(
        private readonly cdn: string,
        private readonly origin: string,
        private readonly root: Wrap<any>,
        private readonly onLocalAuth: (email: string, password: string) => Promise<void>,
        private readonly onOAuth: (idp: string) => Promise<void>,
        private readonly onSignup: (name: string, website: string, email: string, password: string) => Promise<void>,
    ) {
        super(UIToolkit.div('profile-bar').element);
    }

    /**
     * Map of allowed authentication methods: {idp: true}.
     */
    set authMethods(am: StringBooleanMap) {
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
        this.btnLogin = null;

        // Create an avatar element
        const idxColor = Utils.colourIndex(`${commenter.commenterHex}-${commenter.name}`);
        const avatar = commenter.photo === 'undefined' ?
            UIToolkit.div('avatar', `bg-${idxColor}`).html(commenter.name[0].toUpperCase()) :
            Wrap.new('img')
                .classes('avatar-img')
                .attr({
                    src: `${this.cdn}/api/commenter/photo?commenterHex=${commenter.commenterHex}`,
                    loading: 'lazy',
                    alt: '',
                });

        // Recreate the content
        const link = !commenter.link || commenter.link === 'undefined' ? undefined : commenter.link;
        this.html('')
            .append(
                // Commenter avatar and name
                UIToolkit.div('logged-in-as')
                    .append(
                        // Avatar
                        avatar,
                        // Name and link
                        Wrap.new(link ? 'a' : 'div')
                            .classes('name')
                            .inner(commenter.name)
                            .attr({href: link, rel: link && 'nofollow noopener noreferrer'})),
                // Buttons on the right
                UIToolkit.div()
                    .append(
                        // If it's a local user, add a Profile link
                        commenter.provider === 'commento' &&
                        Wrap.new('a')
                            .classes('profile-link')
                            .inner('Profile')
                            .attr({href: `${this.origin}/profile?commenterToken=${token}`, target: '_blank'}),
                        // Notifications link
                        Wrap.new('a')
                            .classes('profile-link')
                            .inner('Notifications')
                            .attr({
                                href: `${this.origin}/unsubscribe?unsubscribeSecretHex=${email.unsubscribeSecretHex}`,
                                target: '_blank',
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
            {ref: this.btnLogin, placement: 'bottom-end'},
            this._authMethods,
            this.origin);
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
        const dlg = await SignupDialog.run(this.root, {ref: this.btnLogin, placement: 'bottom-end'});
        return dlg.confirmed && await this.onSignup(dlg.name, dlg.website, dlg.email, dlg.password);
    }
}
