import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:path_provider/path_provider.dart';

class DebugLogsView extends StatefulWidget {
  @override
  _DebugLogsViewState createState() => _DebugLogsViewState();
}

class _DebugLogsViewState extends State<DebugLogsView> {
  String _logs = 'Loading logs...';
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadLogs();
  }

  Future<void> _loadLogs() async {
    try {
      final directory = await getApplicationDocumentsDirectory();
      final file = File('${directory.path}/sync_debug.log');

      if (await file.exists()) {
        final content = await file.readAsString();
        setState(() {
          _logs = content.isEmpty ? 'No logs found' : content;
          _isLoading = false;
        });
      } else {
        setState(() {
          _logs = 'Log file does not exist yet. Trigger some sync operations first.';
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() {
        _logs = 'Error loading logs: $e';
        _isLoading = false;
      });
    }
  }

  Future<void> _copyLogsToClipboard() async {
    try {
      await Clipboard.setData(ClipboardData(text: _logs));
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Logs copied to clipboard'))
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Error copying logs: $e'))
      );
    }
  }

  Future<void> _clearLogs() async {
    try {
      final directory = await getApplicationDocumentsDirectory();
      final file = File('${directory.path}/sync_debug.log');

      if (await file.exists()) {
        await file.delete();
      }

      setState(() {
        _logs = 'Logs cleared';
      });
    } catch (e) {
      setState(() {
        _logs = 'Error clearing logs: $e';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Debug Logs'),
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: () {
              setState(() {
                _isLoading = true;
              });
              _loadLogs();
            },
          ),
          IconButton(
            icon: Icon(Icons.copy),
            onPressed: _copyLogsToClipboard,
          ),
          IconButton(
            icon: Icon(Icons.clear),
            onPressed: _clearLogs,
          ),
        ],
      ),
      body: _isLoading
          ? Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: EdgeInsets.all(16),
              child: SelectableText(
                _logs,
                style: TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 12,
                ),
              ),
            ),
    );
  }
}