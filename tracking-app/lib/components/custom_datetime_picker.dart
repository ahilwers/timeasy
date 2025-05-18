import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class CustomDateTimePicker extends StatelessWidget {
  final DateTime? dateTime;
  final String? dateHint;
  final String? timeHint;
  final Function(DateTime) onDateTimeChanged;
  final bool showBorder;
  final Color? accentColor;

  CustomDateTimePicker({
    required this.dateTime,
    required this.onDateTimeChanged,
    this.dateHint,
    this.timeHint,
    this.showBorder = true,
    this.accentColor,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final primaryColor = accentColor ?? theme.colorScheme.primary;
    final textColor = isDark ? Colors.white : Colors.black87;
    final localizations = AppLocalizations.of(context)!;
    
    // Format the date and time
    final locale = Localizations.localeOf(context).toString();
    final dateFormatter = DateFormat.yMd(locale);
    final timeFormatter = DateFormat.Hm(locale);
    
    final dateText = dateTime != null 
        ? dateFormatter.format(dateTime!.toLocal()) 
        : dateHint ?? localizations.endDate;
    
    final timeText = dateTime != null 
        ? timeFormatter.format(dateTime!.toLocal()) 
        : timeHint ?? localizations.endTime;

    // Get the input decoration theme to match text fields
    final inputDecorationTheme = theme.inputDecorationTheme;
    final borderRadius = inputDecorationTheme.border is OutlineInputBorder
        ? (inputDecorationTheme.border as OutlineInputBorder).borderRadius
        : BorderRadius.circular(4.0);
    
    // Use the theme's fill color or a default
    final fillColor = inputDecorationTheme.fillColor ?? 
        (isDark ? Colors.grey[800] : Colors.grey[100]);
    
    // Use the theme's border color or a default
    final borderColor = inputDecorationTheme.border is OutlineInputBorder
        ? (inputDecorationTheme.border as OutlineInputBorder).borderSide.color
        : theme.dividerColor;

    return Container(
      margin: EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          // Date picker button
          Expanded(
            child: _buildPickerButton(
              context,
              Icons.calendar_today,
              dateText,
              primaryColor,
              textColor,
              () => _selectDate(context),
              showLeftBorder: showBorder,
              showRightBorder: showBorder,
              borderRadius: borderRadius,
              fillColor: fillColor,
              borderColor: borderColor,
              isDark: isDark,
            ),
          ),
          // Add spacing between date and time pickers
          SizedBox(width: 8),
          // Time picker button
          Expanded(
            child: _buildPickerButton(
              context,
              Icons.access_time,
              timeText,
              primaryColor,
              textColor,
              () => _selectTime(context),
              showLeftBorder: showBorder,
              showRightBorder: showBorder,
              borderRadius: borderRadius,
              fillColor: fillColor,
              borderColor: borderColor,
              isDark: isDark,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPickerButton(
    BuildContext context,
    IconData icon,
    String text,
    Color primaryColor,
    Color textColor,
    VoidCallback onPressed,
    {
      bool showLeftBorder = true, 
      bool showRightBorder = true,
      BorderRadius? borderRadius,
      Color? fillColor,
      Color? borderColor,
      bool isDark = false,
    }
  ) {
    return InkWell(
      onTap: onPressed,
      borderRadius: borderRadius ?? BorderRadius.circular(4),
      child: Container(
        padding: EdgeInsets.symmetric(vertical: 16, horizontal: 16),
        decoration: BoxDecoration(
          color: fillColor ?? (isDark ? Colors.grey[800] : Colors.grey[50]),
          borderRadius: borderRadius ?? BorderRadius.horizontal(
            left: showLeftBorder ? Radius.circular(4) : Radius.zero,
            right: showRightBorder ? Radius.circular(4) : Radius.zero,
          ),
          border: Border.all(
            color: borderColor ?? primaryColor.withOpacity(0.3),
            width: 1,
          ),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              color: primaryColor,
              size: 20,
            ),
            SizedBox(width: 12),
            Flexible(
              child: Text(
                text,
                style: TextStyle(
                  color: textColor,
                  fontWeight: FontWeight.w500,
                  fontSize: 16,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _selectDate(BuildContext context) async {
    final initialDate = dateTime ?? DateTime.now();
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: DateTime(2010),
      lastDate: DateTime(2201),
      builder: (BuildContext context, Widget? child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: Theme.of(context).colorScheme.copyWith(
              primary: accentColor ?? Theme.of(context).colorScheme.primary,
            ),
          ),
          child: child!,
        );
      },
    );
    
    if (picked != null) {
      final newDateTime = _combineDateAndTime(
        picked, 
        dateTime != null ? TimeOfDay.fromDateTime(dateTime!.toLocal()) : TimeOfDay.now()
      );
      onDateTimeChanged(newDateTime);
    }
  }

  Future<void> _selectTime(BuildContext context) async {
    final initialTime = dateTime != null 
        ? TimeOfDay.fromDateTime(dateTime!.toLocal()) 
        : TimeOfDay.now();
        
    final TimeOfDay? picked = await showTimePicker(
      context: context,
      initialTime: initialTime,
      builder: (BuildContext context, Widget? child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: Theme.of(context).colorScheme.copyWith(
              primary: accentColor ?? Theme.of(context).colorScheme.primary,
              onPrimary: Theme.of(context).brightness == Brightness.dark 
                  ? Colors.white 
                  : Colors.black,
            ),
            timePickerTheme: TimePickerThemeData(
              backgroundColor: Theme.of(context).scaffoldBackgroundColor,
              hourMinuteShape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              dayPeriodShape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              dayPeriodColor: MaterialStateColor.resolveWith((states) => 
                states.contains(MaterialState.selected) 
                  ? (accentColor ?? Theme.of(context).colorScheme.primary).withOpacity(0.12)
                  : Colors.transparent
              ),
              dayPeriodTextColor: MaterialStateColor.resolveWith((states) => 
                states.contains(MaterialState.selected)
                  ? (accentColor ?? Theme.of(context).colorScheme.primary)
                  : Theme.of(context).colorScheme.onSurface
              ),
              hourMinuteColor: MaterialStateColor.resolveWith((states) => 
                states.contains(MaterialState.selected)
                  ? (accentColor ?? Theme.of(context).colorScheme.primary)
                  : Theme.of(context).colorScheme.surface
              ),
              hourMinuteTextColor: MaterialStateColor.resolveWith((states) => 
                states.contains(MaterialState.selected)
                  ? Theme.of(context).colorScheme.onPrimary
                  : Theme.of(context).colorScheme.onSurface
              ),
            ),
          ),
          child: child!,
        );
      },
    );
    
    if (picked != null) {
      final newDateTime = _combineDateAndTime(
        dateTime != null ? dateTime!.toLocal() : DateTime.now(),
        picked
      );
      onDateTimeChanged(newDateTime);
    }
  }

  DateTime _combineDateAndTime(DateTime date, TimeOfDay time) {
    return DateTime(
      date.year,
      date.month,
      date.day,
      time.hour,
      time.minute,
    ).toUtc();
  }
}
