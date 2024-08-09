import 'dart:io';

import 'package:excel/excel.dart';
import 'package:flutter/material.dart';
import 'package:timeasy/models/timeentry.dart';
import 'package:timeasy/repositories/timeentry_repository.dart';
import 'package:timeasy/tools/date_tools.dart';

class ExcelExport {
  final String directory;
  final DateTimeRange dateRange;
  final String projectId;

  ExcelExport(this.directory, this.dateRange, this.projectId);

  Future<void> Export() async {
    var dateTools = DateTools();
    var timeEntries = await getTimeEntries();
    var excel = Excel.createExcel();
    var defautSheetName = excel.getDefaultSheet() ?? "";
    var sheetObject = excel[defautSheetName];
    addHeader(sheetObject);
    var dayTimeEntries = List<TimeEntry>.empty(growable: true);
    DateTime? lastDate = null;
    var currentLine = 2;
    for (var i = 0; i < timeEntries.length; i++) {
      var startDate = dateTools.onlyDate(timeEntries[i].startTime);
      if ((lastDate == null) || (startDate != lastDate)) {
        addTimeEntriesToSheet(sheetObject, i + 1, dayTimeEntries);
        dayTimeEntries.clear();
        currentLine++;
      }
      dayTimeEntries.add(timeEntries[i]);
      lastDate = dateTools.onlyDate(timeEntries[i].startTime);
    }
    currentLine++;
    addTimeEntriesToSheet(sheetObject, currentLine, dayTimeEntries);
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
    sheetObject.cell(CellIndex.indexByString("D1")).value =
        TextCellValue("Pause 1 Start");
    sheetObject.cell(CellIndex.indexByString("E1")).value =
        TextCellValue("Pause 1 End");
    sheetObject.cell(CellIndex.indexByString("F1")).value =
        TextCellValue("Pause 2 Start");
    sheetObject.cell(CellIndex.indexByString("G1")).value =
        TextCellValue("Pause 2 End");
    sheetObject.cell(CellIndex.indexByString("H1")).value =
        TextCellValue("Pause 3 Start");
    sheetObject.cell(CellIndex.indexByString("I1")).value =
        TextCellValue("Pause 3 End");
    sheetObject.cell(CellIndex.indexByString("J1")).value =
        TextCellValue("Pause 4 Start");
    sheetObject.cell(CellIndex.indexByString("K1")).value =
        TextCellValue("Pause 4 End");
    sheetObject.cell(CellIndex.indexByString("L1")).value =
        TextCellValue("Pause 5 Start");
    sheetObject.cell(CellIndex.indexByString("M1")).value =
        TextCellValue("Pause 5 End");
  }

  void addTimeEntriesToSheet(
      Sheet sheetObject, int currentLine, List<TimeEntry> timeEntries) {
    if (timeEntries.isEmpty) {
      return;
    }
    var dateCell = sheetObject.cell(CellIndex.indexByString("A${currentLine}"));
    dateCell.value = TextCellValue(formatDate(timeEntries[0].startTime));
    var startTimeCell =
        sheetObject.cell(CellIndex.indexByString("B${currentLine}"));
    startTimeCell.value = TextCellValue(formatTime(timeEntries[0].startTime));
    var endTimeCell =
        sheetObject.cell(CellIndex.indexByString("C${currentLine}"));
    endTimeCell.value = timeEntries[timeEntries.length - 1].endTime == null
        ? TextCellValue("")
        : TextCellValue(
            formatTime(timeEntries[timeEntries.length - 1].endTime!));
    var pauseCellIndex = 65 + 3; // D
    for (var i = 0; i < timeEntries.length - 1; i++) {
      var timeEntry = timeEntries[i];
      if (timeEntry.endTime == null) {
        continue;
      }
      var pauseStartCell = sheetObject.cell(CellIndex.indexByString(
          "${String.fromCharCode(pauseCellIndex)}${currentLine}"));
      pauseStartCell.value = TextCellValue(formatTime(timeEntry.endTime!));
      pauseCellIndex++;
      var nextTimeEntry = timeEntries[i + 1];
      var pauseEndCell = sheetObject.cell(CellIndex.indexByString(
          "${String.fromCharCode(pauseCellIndex)}${currentLine}"));
      pauseEndCell.value = TextCellValue(formatTime(nextTimeEntry.startTime));
      pauseCellIndex++;
    }
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
