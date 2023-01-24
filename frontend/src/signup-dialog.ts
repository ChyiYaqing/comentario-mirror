import { Wrap } from './element-wrap';

export enum SignupStatus {
    /** User dismissed the dialog. */
    CANCEL,
    /** User confirmed (submitted) the dialog. */
    OK,
}

/**
 * The outcome of the signup dialog.
 */
export interface SignupResult {
    readonly status:   SignupStatus;
    readonly name:     string;
    readonly website:  string;
    readonly email:    string;
    readonly password: string;
}

export class SignupDialog {

    private backdrop: Wrap<HTMLDivElement>;
    private container: Wrap<HTMLDivElement>;
    private dialogBox: Wrap<HTMLFormElement>;
    private nameInput: Wrap<HTMLInputElement>;
    private websiteInput: Wrap<HTMLInputElement>;
    private emailInput: Wrap<HTMLInputElement>;
    private passwordInput: Wrap<HTMLInputElement>;

    constructor(
        private readonly parent: Wrap<any>,
        private readonly resolve: (r: SignupResult | PromiseLike<SignupResult>) => void,
    ) {}

    /**
     * Instantiate and show the dialog. Return a promise that resolves as soon as the dialog is closed, with the outcome
     * of the operation.
     * @param parent Parent element for the dialog.
     */
    static run(parent: Wrap<any>): Promise<SignupResult> {
        return new Promise(resolve => new SignupDialog(parent, resolve).render());
    }

    private render() {
        // Create a login box
        this.dialogBox = Wrap.new('form')
            .classes('login-box', 'fade-in')
            // Form submit event
            .on('submit', e => {
                e.preventDefault();
                this.dismiss(SignupStatus.OK);
            })
            // Don't propagate the click to prevent cancelling the dialog, which happens when the click reaches the
            // parent container
            .click(e => e.stopPropagation())
            // Close button
            .append(
                Wrap.new('button')
                    .classes('btn-login-box-close')
                    .attr({type: 'button', ariaLabel: 'Close'})
                    .click(() => this.dismiss(SignupStatus.CANCEL)));

        // Create inputs
        this.nameInput = Wrap.new('input')
            .classes('input')
            .attr({name: 'name', placeholder: 'Real name', type: 'text', autocomplete: 'name'});
        this.websiteInput = Wrap.new('input')
            .classes('input')
            .attr({name: 'website', placeholder: 'Website (optional)', type: 'text', autocomplete: 'url'});
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
                .inner('Create an account'),
            // Name input container
            Wrap.new('div')
                .classes('email-container')
                .append(Wrap.new('div').classes('email').append(this.nameInput)),
            // Website input container
            Wrap.new('div')
                .classes('email-container')
                .append(Wrap.new('div').classes('email').append(this.websiteInput)),
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
                            Wrap.new('button').classes('email-button').inner('Sign up').attr({type: 'submit'}))));

        // Create a backdrop
        this.backdrop = Wrap.new('div').classes('backdrop').appendTo(this.parent);

        // Add the dialog to the container and scroll to it, if necessary
        this.container = Wrap.new('div')
            .classes('login-box-container')
            // Cancel the dialog when clicked outside
            .click(() => this.dismiss(SignupStatus.CANCEL))
            .appendTo(this.parent);
        this.dialogBox.appendTo(this.container).scrollTo();

        // Focus the email input
        this.nameInput.focus();
    }

    private dismiss(status: SignupStatus) {
        const name = this.nameInput.val;
        const website = this.websiteInput.val;
        const email = this.emailInput.val;
        const password = this.passwordInput.val;

        // Close the dialog
        this.dialogBox.noClasses('fade-in').classes('fade-out');
        setTimeout(
            () => {
                this.container.remove();
                this.backdrop.remove();

                // Resolve the promise, returning the result
                this.resolve({status, name, website, email, password});
            },
            250);
    }
}
