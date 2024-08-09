import 'dart:io';

import 'package:excel/excel.dart';
import 'package:flutter/material.dart';
import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';

class ExcelExportAllEntries {
  final String directory;
  final DateTimeRange dateRange;
  final String projectId;

  ExcelExportAllEntries(this.directory, this.dateRange, this.projectId);

  Future<void> Export() async {
    var timeEntries = await getTimeEntries();
    var excel = Excel.createExcel();
    var defautSheetName = excel.getDefaultSheet() ?? "";
    var sheetObject = excel[defautSheetName];
    addHeader(sheetObject);
    var currentLine = 2;
    timeEntries.forEach((timeEntry) {
      addTimeEntryToSheet(sheetObject, currentLine, timeEntry);
      currentLine++;
    });
    var fileBytes = excel.save();

    File('$directory/output_file_name.xlsx')
      ..createSync(recursive: true)
      ..writeAsBytesSync(fileBytes!);
  }

  void addHeader(Sheet sheetObject) {
    sheetObject.cell(CellIndex.indexByString("A1")).value =
        TextCellValue("Datum");
    sheetObject.cell(CellIndex.indexByString("B1")).value =
        TextCellValue("Start Time");
    sheetObject.cell(CellIndex.indexByString("C1")).value =
        TextCellValue("End Time");
  }

  void addTimeEntryToSheet(
      Sheet sheetObject, int currentLine, TimeEntry timeEntry) {
    var dateCell = sheetObject.cell(CellIndex.indexByString("A${currentLine}"));
    dateCell.value = TextCellValue(formatDate(timeEntry.startTime));
    var startTimeCell =
        sheetObject.cell(CellIndex.indexByString("B${currentLine}"));
    startTimeCell.value = TextCellValue(formatTime(timeEntry.startTime));
    var endTimeCell =
        sheetObject.cell(CellIndex.indexByString("C${currentLine}"));
    endTimeCell.value = timeEntry.endTime == null
        ? TextCellValue("")
        : TextCellValue(formatTime(timeEntry.endTime!));
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
    return "${localTime.hour.toString().padLeft(2, '0')}:${localTime.minute.toString().padLeft(2, '0')}";
  }
}
