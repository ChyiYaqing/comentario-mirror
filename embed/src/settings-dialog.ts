import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { Dialog, DialogPositioning } from './dialog';
import { Commenter, ProfileSettings, Email } from './models';

export class SettingsDialog extends Dialog {

    private _name?: Wrap<HTMLInputElement>;
    private _website?: Wrap<HTMLInputElement>;
    private _email?: Wrap<HTMLInputElement>;
    private _avatar?: Wrap<HTMLInputElement>;
    private _cbNotifyModerator?: Wrap<HTMLInputElement>;
    private _cbNotifyReplies?: Wrap<HTMLInputElement>;

    private constructor(parent: Wrap<any>, pos: DialogPositioning, private readonly commenter: Commenter, private readonly email: Email) {
        super(parent, 'Profile settings', pos);
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     * @param pos Positioning options.
     * @param commenter Commenter whose profile settings are being edited.
     * @param email Email that defines notification settings.
     */
    static run(parent: Wrap<any>, pos: DialogPositioning, commenter: Commenter, email: Email): Promise<SettingsDialog> {
        const dlg = new SettingsDialog(parent, pos, commenter, email);
        return dlg.run(dlg);
    }

    /**
     * Entered settings.
     */
    get data(): ProfileSettings {
        return {
            email:           this._email?.val   || '',
            name:            this._name?.val    || '',
            websiteUrl:      this._website?.val || '',
            avatarUrl:       this._avatar?.val  || '',
            notifyModerator: !!this._cbNotifyModerator?.isChecked,
            notifyReplies:   !!this._cbNotifyReplies?.isChecked,
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
                // Checkboxes
                UIToolkit.div('checkbox-group').append(
                    // Moderator notifications checkbox (only if the current commenter is a moderator)
                    this.commenter.isModerator && UIToolkit.div('checkbox-container')
                        .append(
                            this._cbNotifyModerator = Wrap.new('input')
                                .id('cb-notify-moderator')
                                .attr({type: 'checkbox'})
                                .checked(!!this.email.sendModeratorNotifications),
                            Wrap.new('label').attr({for: this._cbNotifyModerator.getAttr('id')}).inner('Moderator notifications')),
                    // Reply notifications checkbox
                    UIToolkit.div('checkbox-container')
                        .append(
                            this._cbNotifyReplies = Wrap.new('input')
                                .id('cb-notify-replies')
                                .attr({type: 'checkbox'})
                                .checked(!!this.email.sendReplyNotifications),
                            Wrap.new('label').attr({for: this._cbNotifyReplies.getAttr('id')}).inner('Reply notifications'))),
                // Submit button
                UIToolkit.div('dialog-centered').append(UIToolkit.submit('Save', false)));
    }

    override onShow(): void {
        this._email?.focus();
    }
}
