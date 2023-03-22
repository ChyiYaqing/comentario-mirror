import { Language } from '../app/_models/models';

// Available UI languages
export const languages: Language[] = [
    // Work around the strange choice of 2-digit year for the default en-US locale
    {nativeName: 'English', code: 'en', weight: 10, dateFormat: 'M/d/yyyy', datetimeFormat: 'M/d/yyyy, h:mm a'},
];
