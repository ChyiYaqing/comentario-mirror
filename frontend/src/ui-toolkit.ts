import { Wrap } from './element-wrap';

/**
 * Utility class to facilitate the creation of various UI components.
 */
export class UIToolkit {

    /**
     * Create and return a dialog close button.
     * @param onClick Button's click handler.
     */
    static closeButton(onClick: (e: MouseEvent) => void): Wrap<HTMLButtonElement> {
        return Wrap.new('button')
            .classes('dialog-btn-close')
            .attr({type: 'button', ariaLabel: 'Close'})
            .click(onClick);
    }

    /**
     * Create and return a new popup dialog element.
     * @param onSubmit Form submit handler
     */
    static form(onSubmit: () => void): Wrap<HTMLFormElement> {
        return Wrap.new('form')
            // Form submit event
            .on('submit', e => {
                e.preventDefault();
                onSubmit();
            });
    }

    /**
     * Create and return a new input element.
     */
    static input(name: string, type = 'text', placeholder: string = null, autocomplete: string = null): Wrap<HTMLInputElement> {
        return Wrap.new('input').classes('input').attr({name, type, placeholder, autocomplete});
    }
}
