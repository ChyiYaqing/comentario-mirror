import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { Dialog, DialogPositioning } from './dialog';
import { Commenter, CommenterSettings } from './models';

export class SettingsDialog extends Dialog {

    private _name?: Wrap<HTMLInputElement>;
    private _website?: Wrap<HTMLInputElement>;
    private _email?: Wrap<HTMLInputElement>;
    private _avatar?: Wrap<HTMLInputElement>;

    private constructor(parent: Wrap<any>, pos: DialogPositioning, private readonly commenter: Commenter) {
        super(parent, 'Profile settings', pos);
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     * @param pos Positioning options..
     * @param commenter Commenter whose profile settings are being edited.
     */
    static run(parent: Wrap<any>, pos: DialogPositioning, commenter: Commenter): Promise<SettingsDialog> {
        const dlg = new SettingsDialog(parent, pos, commenter);
        return dlg.run(dlg);
    }

    /**
     * Entered settings.
     */
    get data(): CommenterSettings {
        return {
            email:      this._email?.val   || '',
            name:       this._name?.val    || '',
            websiteUrl: this._website?.val || '',
            avatarUrl:  this._avatar?.val  || '',
        };
    }

    override renderContent(): Wrap<any> {
        // Create inputs if it's a local user
        const inputs: Wrap<any>[] = [];
        if (!this.commenter.provider) {
            this._email   = UIToolkit.input('email',   'email', 'Email address', 'email', true).value(this.commenter.email      || '');
            this._name    = UIToolkit.input('name',    'text',  'Real name',     'name', true) .value(this.commenter.name       || '');
            this._website = UIToolkit.input('website', 'url',   'Website',       'url')        .value(this.commenter.websiteUrl || '');
            this._avatar  = UIToolkit.input('avatar',  'url',   'Avatar URL')                  .value(this.commenter.avatarUrl  || '');
            inputs.push(
                UIToolkit.div('input-group').append(this._email),
                UIToolkit.div('input-group').append(this._name),
                UIToolkit.div('input-group').append(this._website),
                UIToolkit.div('input-group').append(this._avatar));
        }

        // Add the inputs to a new form
        return UIToolkit.form(() => this.dismiss(true), () => this.dismiss())
            .append(
                ...inputs,
                UIToolkit.div('dialog-centered').append(UIToolkit.submit('Save', false)));
    }

    override onShow(): void {
        this._email?.focus();
    }
}
