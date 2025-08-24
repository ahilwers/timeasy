import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslatePipe } from '@ngx-translate/core';

// PrimeNG imports
import { Dialog } from 'primeng/dialog';
import { Button } from 'primeng/button';
import { Divider } from 'primeng/divider';

@Component({
  selector: 'app-token-documentation',
  standalone: true,
  imports: [
    CommonModule,
    TranslatePipe,
    Dialog,
    Button,
    Divider
  ],
  templateUrl: './token-documentation.component.html',
  styleUrl: './token-documentation.component.css'
})
export class TokenDocumentationComponent {
  @Input() visible = false;
  @Input() provider: string = '';
  @Output() visibleChange = new EventEmitter<boolean>();

  onDialogHide() {
    this.visible = false;
    this.visibleChange.emit(false);
  }

  getDialogTitle(): string {
    return `tokenDocumentation.${this.provider}.title`;
  }

  getSteps(): string[] {
    const steps: string[] = [];
    const maxSteps = this.getMaxSteps();
    
    for (let i = 1; i <= maxSteps; i++) {
      steps.push(`tokenDocumentation.${this.provider}.step${i}`);
    }
    
    return steps;
  }

  private getMaxSteps(): number {
    switch (this.provider) {
      case 'github':
        return 8;
      case 'gitlab':
        return 8;
      case 'jira':
        return 7;
      default:
        return 0;
    }
  }

  getPermissionsList(): string[] {
    const permissions: string[] = [];
    const maxPermissions = this.getMaxPermissions();
    
    for (let i = 1; i <= maxPermissions; i++) {
      permissions.push(`tokenDocumentation.${this.provider}.permission${i}`);
    }
    
    return permissions;
  }

  private getMaxPermissions(): number {
    switch (this.provider) {
      case 'github':
        return 3;
      case 'gitlab':
        return 3;
      case 'jira':
        return 2;
      default:
        return 0;
    }
  }
}