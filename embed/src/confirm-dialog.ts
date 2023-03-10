import { Wrap } from './element-wrap';
import { Dialog, DialogPositioning } from './dialog';
import { UIToolkit } from './ui-toolkit';

export class ConfirmDialog extends Dialog {

    private btnOk?: Wrap<HTMLButtonElement>;

    private constructor(parent: Wrap<any>, pos: DialogPositioning, private readonly text: string) {
        super(parent, 'Confirm', pos);
    }

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed.
     * @param parent Parent element for the dialog.
     * @param pos Positioning options.
     * @param text Dialog text.
     */
    static run(parent: Wrap<any>, pos: DialogPositioning, text: string): Promise<boolean> {
        const dlg = new ConfirmDialog(parent, pos, text);
        return dlg.run(null).then(() => dlg.confirmed);
    }

    override renderContent(): Wrap<any> {
        this.btnOk = UIToolkit.button('OK', () => this.dismiss(true), 'danger-button');
        return UIToolkit.div()
            .append(
                // Dialog text
                UIToolkit.div('dialog-centered').inner(this.text),
                // Button
                UIToolkit.div('dialog-centered').append(UIToolkit.button('Cancel', () => this.dismiss()), this.btnOk));
    }

    override onShow() {
        this.btnOk?.focus();
    }
}
