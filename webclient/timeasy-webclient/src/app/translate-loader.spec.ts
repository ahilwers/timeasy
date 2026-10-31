import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TRANSLATE_HTTP_LOADER_CONFIG, TranslateHttpLoader } from '@ngx-translate/http-loader';

describe('TranslateHttpLoader', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        TranslateHttpLoader,
        {
          provide: TRANSLATE_HTTP_LOADER_CONFIG,
          useValue: { resources: [{ prefix: './i18n/', suffix: '.json' }] },
        },
      ],
    });
  });

  it('requests the translation files served next to the app', () => {
    const loader = TestBed.inject(TranslateHttpLoader);
    const httpMock = TestBed.inject(HttpTestingController);

    let translations: object | undefined;
    loader.getTranslation('de').subscribe((result) => (translations = result));

    httpMock.expectOne('./i18n/de.json').flush({ globals: { yes: 'Ja' } });

    expect(translations).toEqual({ globals: { yes: 'Ja' } });
    httpMock.verify();
  });

  it('returns no translations when no resources are configured', () => {
    // Guards against a misconfigured loader silently showing translation keys.
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        TranslateHttpLoader,
        { provide: TRANSLATE_HTTP_LOADER_CONFIG, useValue: {} },
      ],
    });

    const loader = TestBed.inject(TranslateHttpLoader);

    let translations: object | undefined;
    loader.getTranslation('de').subscribe((result) => (translations = result));

    expect(translations).toEqual({});
  });
});
