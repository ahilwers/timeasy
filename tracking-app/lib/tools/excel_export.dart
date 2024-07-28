import 'dart:io';

import 'package:excel/excel.dart';
import 'package:flutter/material.dart';
import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';

class ExcelExport {
  final String directory;
  final DateTimeRange dateRange;
  final String projectId;

  ExcelExport(this.directory, this.dateRange, this.projectId);

  Future<void> Export() async {
    var timeEntries = await getTimeEntries();
    var excel = Excel.createExcel();
    var defautSheetName = excel.getDefaultSheet() ?? "";
    var sheetObject = excel[defautSheetName];
    for (var i = 0; i < timeEntries.length; i++) {
      var startTimeCell =
          sheetObject.cell(CellIndex.indexByString("A${i + 1}"));
      startTimeCell.value = TextCellValue(formatTime(timeEntries[i].startTime));
      var endTimeCell = sheetObject.cell(CellIndex.indexByString("B${i + 1}"));
      endTimeCell.value = timeEntries[i].endTime == null
          ? TextCellValue("")
          : TextCellValue(formatTime(timeEntries[i].endTime!));
      var descriptionCell =
          sheetObject.cell(CellIndex.indexByString("C${i + 1}"));
      var description = timeEntries[i].description ?? "";
      descriptionCell.value = TextCellValue(description);
    }
    var fileBytes = excel.save();

    File('$directory/output_file_name.xlsx')
      ..createSync(recursive: true)
      ..writeAsBytesSync(fileBytes!);
  }

  Future<List<TimeEntry>> getTimeEntries() async {
    return await TimeEntryRepository()
        .getTimeEntries(projectId, dateRange.start, dateRange.end);
  }

  String formatDate(DateTime date) {
    DateTime localDate = date.toLocal();
    return "${localDate.year}-${localDate.month.toString().padLeft(2, '0')}-${localDate.day.toString().padLeft(2, '0')}";
  }

  String formatTime(DateTime time) {
    DateTime localTime = time.toLocal();
    return "${localTime.hour.toString().padLeft(2, '0')}:${localTime.minute.toString().padLeft(2, '0')}:${localTime.second.toString().padLeft(2, '0')}";
  }
}
