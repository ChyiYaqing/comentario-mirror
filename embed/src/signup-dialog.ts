import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { Dialog, DialogPositioning } from './dialog';
import { SignupData } from './models';

export class SignupDialog extends Dialog {

    private _name?: Wrap<HTMLInputElement>;
    private _website?: Wrap<HTMLInputElement>;
    private _email?: Wrap<HTMLInputElement>;
    private _pwd?: Wrap<HTMLInputElement>;

    private constructor(parent: Wrap<any>, pos: DialogPositioning) {
        super(parent, 'Create an account', pos);
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     * @param pos Positioning options..
     */
    static run(parent: Wrap<any>, pos: DialogPositioning): Promise<SignupDialog> {
        const dlg = new SignupDialog(parent, pos);
        return dlg.run(dlg);
    }

    /**
     * Entered data.
     */
    get data(): SignupData {
        return {
            email:      this._email?.val   || '',
            name:       this._name?.val    || '',
            password:   this._pwd?.val     || '',
            websiteUrl: this._website?.val || '',
        };
    }

    override renderContent(): Wrap<any> {
        // Create inputs
        this._email   = UIToolkit.input('email',    'email',    'Email address',      'email', true);
        this._name    = UIToolkit.input('name',     'text',     'Real name',          'name', true);
        this._pwd     = UIToolkit.input('password', 'password', 'Password',           'current-password', true);
        this._website = UIToolkit.input('website',  'url',      'Website (optional)', 'url');

        // Add the inputs to a new form
        return UIToolkit.form(() => this.dismiss(true), () => this.dismiss())
            .append(
                UIToolkit.div('input-group').append(this._email),
                UIToolkit.div('input-group').append(this._name),
                UIToolkit.div('input-group').append(this._pwd),
                UIToolkit.div('input-group').append(this._website),
                UIToolkit.div('dialog-centered').append(UIToolkit.submit('Sign up', false)));
    }

    override onShow(): void {
        this._email?.focus();
    }
}
