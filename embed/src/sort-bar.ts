import { Wrap } from './element-wrap';
import { UIToolkit } from './ui-toolkit';
import { sortingProps, SortPolicy } from './models';

export class SortBar extends Wrap<HTMLDivElement> {

    private readonly buttons: { sp: SortPolicy; btn: Wrap<HTMLAnchorElement>; }[] = [];

    constructor(
        private readonly onChange: (sp: SortPolicy) => void,
        initialSort: SortPolicy,
    ) {
        super(UIToolkit.div('sort-policy-buttons-container').element);

        // Create sorting buttons
        const cont = UIToolkit.div('sort-policy-buttons').appendTo(this);
        Object.keys(sortingProps).forEach(sp =>
            this.buttons.push({
                sp: sp as SortPolicy,
                btn: Wrap.new('a')
                    .classes('sort-policy-button')
                    .inner(sortingProps[sp as SortPolicy].label)
                    .click(() => this.setSortPolicy(sp as SortPolicy, true))
                    .appendTo(cont),
            }));

        // Apply the initial sorting selection
        this.setSortPolicy(initialSort, false);
    }

    private setSortPolicy(sp: SortPolicy, callOnChange: boolean) {
        this.buttons.forEach(b => b.btn.setClasses(b.sp === sp, 'sort-policy-button-selected'));
        if (callOnChange) {
            this.onChange(sp);
        }
    }
}
