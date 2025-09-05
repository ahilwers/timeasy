import { Component, Input, Output, EventEmitter, forwardRef, inject } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { InputText } from 'primeng/inputtext';
import { ExternalIntegrationService } from '../../../services/external-integration.service';

@Component({
  selector: 'app-description-autocomplete',
  standalone: true,
  imports: [
    CommonModule,
    InputText
  ],
  templateUrl: './description-autocomplete.component.html',
  styleUrl: './description-autocomplete.component.css',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DescriptionAutocompleteComponent),
      multi: true
    }
  ]
})
export class DescriptionAutocompleteComponent implements ControlValueAccessor {
  @Input() projectId?: string;
  @Input() placeholder: string = 'Enter description or reference issues like #123';
  @Input() disabled: boolean = false;

  @Output() suggestionSelected = new EventEmitter<string>();

  private externalService = inject(ExternalIntegrationService);

  value: string = '';
  filteredSuggestions: string[] = [];
  highlightedIndex: number = -1;

  private onChange = (value: string) => {};
  onTouched = () => {};

  onInput(event: Event) {
    const input = event.target as HTMLInputElement;
    const query = input.value;
    this.value = query;
    this.onChange(query);

    if (query.length < 2) {
      this.filteredSuggestions = [];
      this.highlightedIndex = -1;
      return;
    }

    if (!this.projectId) {
      this.filteredSuggestions = [];
      this.highlightedIndex = -1;
      return;
    }

    // Get suggestions from external integration service
    this.externalService.getDescriptionSuggestions(this.projectId, query, 10).subscribe({
      next: (response) => {
        this.filteredSuggestions = response.suggestions || [];
        this.highlightedIndex = -1;
      },
      error: (error) => {
        this.filteredSuggestions = [];
        this.highlightedIndex = -1;
      }
    });
  }

  onKeydown(event: KeyboardEvent) {
    if (this.filteredSuggestions.length === 0) {
      return;
    }

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        this.highlightedIndex = (this.highlightedIndex + 1) % this.filteredSuggestions.length;
        break;
      
      case 'ArrowUp':
        event.preventDefault();
        this.highlightedIndex = this.highlightedIndex <= 0 
          ? this.filteredSuggestions.length - 1 
          : this.highlightedIndex - 1;
        break;
      
      case 'Enter':
        event.preventDefault();
        if (this.highlightedIndex >= 0 && this.highlightedIndex < this.filteredSuggestions.length) {
          this.selectSuggestion(this.filteredSuggestions[this.highlightedIndex]);
        }
        break;
      
      case 'Escape':
        event.preventDefault();
        this.filteredSuggestions = [];
        this.highlightedIndex = -1;
        break;
    }
  }

  selectSuggestion(suggestion: string) {
    this.value = suggestion;
    this.onChange(suggestion);
    this.filteredSuggestions = [];
    this.highlightedIndex = -1;
    this.suggestionSelected.emit(suggestion);
  }

  // ControlValueAccessor implementation
  writeValue(value: string): void {
    this.value = value || '';
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled = isDisabled;
  }
}