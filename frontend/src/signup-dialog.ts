import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { Dialog } from './dialog';

export class SignupDialog extends Dialog {

    private _name: Wrap<HTMLInputElement>;
    private _website: Wrap<HTMLInputElement>;
    private _email: Wrap<HTMLInputElement>;
    private _pwd: Wrap<HTMLInputElement>;

    constructor(parent: Wrap<any>) {
        super(parent);
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     */
    static run(parent: Wrap<any>): Promise<SignupDialog> {
        const dlg = new SignupDialog(parent);
        return dlg.run(dlg);
    }

    /**
     * Entered name.
     */
    get name(): string {
        return this._name.val;
    }

    /**
     * Entered website.
     */
    get website(): string {
        return this._website.val;
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

    override renderContent(): Wrap<any> {
        // Create a login form
        const form = UIToolkit.form(() => this.dismiss(true));

        // Create inputs
        this._name    = UIToolkit.input('name', 'text', 'Real name', 'name');
        this._website = UIToolkit.input('website', 'text', 'Website (optional)', 'url');
        this._email   = UIToolkit.input('email', 'text', 'Email address', 'email');
        this._pwd     = UIToolkit.input('password', 'password', 'Password', 'current-password');

        // Add the inputs to the dialog
        form.append(
            // Subtitle
            Wrap.new('div')
                .classes('login-box-subtitle')
                .inner('Create an account'),
            // Name input container
            Wrap.new('div')
                .classes('input-container')
                .append(Wrap.new('div').classes('input-wrapper').append(this._name)),
            // Website input container
            Wrap.new('div')
                .classes('input-container')
                .append(Wrap.new('div').classes('input-wrapper').append(this._website)),
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
                            Wrap.new('button').classes('input-button').inner('Sign up').attr({type: 'submit'}))));
        return form;
    }

    override onShow(): void {
        this._name.focus();
    }
}
