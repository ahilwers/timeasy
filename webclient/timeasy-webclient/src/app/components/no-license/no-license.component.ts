import { Component } from '@angular/core';
import { TranslateModule } from '@ngx-translate/core';
import { CardComponent } from '../card/card.component';

@Component({
  selector: 'app-no-license',
  standalone: true,
  imports: [TranslateModule, CardComponent],
  templateUrl: './no-license.component.html',
  styleUrl: './no-license.component.css'
})
export class NoLicenseComponent {
}
