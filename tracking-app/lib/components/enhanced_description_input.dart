import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';

import '../bloc/authentication/authentication_bloc.dart';
import '../bloc/authentication/authentication_state.dart';
import '../bloc/internetconnection/internet_connection_bloc.dart';
import '../bloc/internetconnection/internet_connection_state.dart';
import '../environment.dart';
import '../models/external_issue.dart';
import '../services/external_integration_service.dart';

class EnhancedDescriptionInput extends StatefulWidget {
  final String projectId;
  final TextEditingController controller;
  final Function(IssueResolveResult?)? onIssueResolved;
  final Function(String)? onPendingReference;
  final String? initialValue;
  final String? hintText;

  const EnhancedDescriptionInput({
    Key? key,
    required this.projectId,
    required this.controller,
    this.onIssueResolved,
    this.onPendingReference,
    this.initialValue,
    this.hintText,
  }) : super(key: key);

  @override
  _EnhancedDescriptionInputState createState() =>
      _EnhancedDescriptionInputState();
}

class _EnhancedDescriptionInputState extends State<EnhancedDescriptionInput> {
  final ExternalIntegrationService _externalService =
      ExternalIntegrationService();
  final OfflineIssueResolutionService _offlineService =
      OfflineIssueResolutionService();

  Timer? _debounceTimer;
  Timer? _issueDetectionTimer;

  List<String> _suggestions = [];
  IssueResolveResult? _resolvedIssue;
  String? _pendingIssue;
  bool _isResolving = false;

  final FocusNode _focusNode = FocusNode();
  final LayerLink _layerLink = LayerLink();
  OverlayEntry? _overlayEntry;

  @override
  void initState() {
    super.initState();
    if (widget.initialValue != null) {
      widget.controller.text = widget.initialValue!;
    }

    // Initialize the external integration service
    _initializeExternalService();

    widget.controller.addListener(_onTextChanged);
    _focusNode.addListener(_onFocusChanged);
  }

  void _initializeExternalService() {
    final authState = context.read<AuthenticationBloc>().state;
    if (authState is AuthenticationAuthenticated &&
        authState.credentials.accessToken != null) {
      _externalService.initialize(
          Environment.apiBaseUrl, authState.credentials.accessToken!);
    }
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _issueDetectionTimer?.cancel();
    widget.controller.removeListener(_onTextChanged);
    _focusNode.removeListener(_onFocusChanged);
    _focusNode.dispose();
    _removeOverlay();
    super.dispose();
  }

  void _onTextChanged() {
    final text = widget.controller.text;

    // Debounced autocomplete
    _debounceTimer?.cancel();
    _debounceTimer = Timer(Duration(milliseconds: 150), () {
      _getSuggestions(text);
    });

    // Debounced issue detection
    _issueDetectionTimer?.cancel();
    _issueDetectionTimer = Timer(Duration(milliseconds: 300), () {
      _detectAndResolveIssue(text);
    });
  }

  void _onFocusChanged() {
    if (_focusNode.hasFocus && _suggestions.isNotEmpty) {
      _showOverlay();
    } else {
      _removeOverlay();
    }
  }

  Future<void> _getSuggestions(String query) async {
    if (query.length < 2) {
      setState(() {
        _suggestions = [];
      });
      _removeOverlay();
      return;
    }

    try {
      final connectionState = context.read<InternetConnectionBloc>().state;
      bool isOnline = connectionState is InternetConnectionConnected;

      if (isOnline) {
        final response = await _externalService.getDescriptionSuggestions(
          widget.projectId,
          query,
          limit: 5,
        );

        setState(() {
          _suggestions = response.suggestions;
        });

        if (_focusNode.hasFocus && _suggestions.isNotEmpty) {
          _showOverlay();
        }
      } else {
        // Use cached suggestions when offline
        final cachedSuggestions =
            await _externalService.getCachedDescriptionSuggestions(
          widget.projectId,
          query,
        );

        setState(() {
          _suggestions = cachedSuggestions;
        });

        if (_focusNode.hasFocus && _suggestions.isNotEmpty) {
          _showOverlay();
        }
      }
    } catch (e) {
      // Handle error silently for UX
      setState(() {
        _suggestions = [];
      });
      _removeOverlay();
    }
  }

  Future<void> _detectAndResolveIssue(String input) async {
    if (input.trim().isEmpty) {
      _clearIssues();
      return;
    }

    final issuePattern = _externalService.detectIssuePattern(input);
    if (issuePattern == null) {
      _clearIssues();
      return;
    }

    final connectionState = context.read<InternetConnectionBloc>().state;
    bool isOnline = connectionState is InternetConnectionConnected;

    if (isOnline) {
      setState(() {
        _isResolving = true;
      });

      try {
        final result =
            await _externalService.resolveIssue(widget.projectId, input);

        setState(() {
          _isResolving = false;
          if (result.isResolved) {
            _resolvedIssue = result;
            _pendingIssue = null;
          } else {
            _pendingIssue = result.key;
            _resolvedIssue = null;
          }
        });

        if (result.isResolved) {
          widget.onIssueResolved?.call(result);
        }
      } catch (e) {
        setState(() {
          _isResolving = false;
          _pendingIssue = issuePattern.key;
          _resolvedIssue = null;
        });

        _offlineService.addPendingReference(issuePattern.key);
        widget.onPendingReference?.call(issuePattern.key);
      }
    } else {
      // Offline mode - add to pending references
      setState(() {
        _pendingIssue = issuePattern.key;
        _resolvedIssue = null;
      });

      _offlineService.addPendingReference(issuePattern.key);
      widget.onPendingReference?.call(issuePattern.key);
    }
  }

  void _clearIssues() {
    setState(() {
      _resolvedIssue = null;
      _pendingIssue = null;
      _isResolving = false;
    });
    widget.onIssueResolved?.call(null);
  }

  void _showOverlay() {
    _removeOverlay();

    _overlayEntry = OverlayEntry(
      builder: (context) => Positioned(
        width: MediaQuery.of(context).size.width - 32,
        child: CompositedTransformFollower(
          link: _layerLink,
          showWhenUnlinked: false,
          offset: Offset(0.0, 60.0),
          child: Material(
            elevation: 4.0,
            borderRadius: BorderRadius.circular(8.0),
            child: Container(
              constraints: BoxConstraints(maxHeight: 200),
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: _suggestions.length,
                itemBuilder: (context, index) {
                  return ListTile(
                    title: Text(_suggestions[index]),
                    onTap: () {
                      widget.controller.text = _suggestions[index];
                      widget.controller.selection = TextSelection.fromPosition(
                        TextPosition(offset: widget.controller.text.length),
                      );
                      _removeOverlay();
                      _focusNode.unfocus();
                    },
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );

    Overlay.of(context).insert(_overlayEntry!);
  }

  void _removeOverlay() {
    _overlayEntry?.remove();
    _overlayEntry = null;
  }

  Widget _buildIssueChip() {
    if (_resolvedIssue != null) {
      return Container(
        margin: EdgeInsets.only(top: 8),
        child: Chip(
          avatar: Icon(Icons.link, size: 16),
          label:
              Text('${_resolvedIssue!.key} · ${_resolvedIssue!.title ?? ''}'),
          backgroundColor: Colors.green.shade100,
          deleteIcon: Icon(Icons.close, size: 16),
          onDeleted: _clearIssues,
        ),
      );
    }

    if (_pendingIssue != null) {
      return Container(
        margin: EdgeInsets.only(top: 8),
        child: Chip(
          avatar: Icon(Icons.schedule, size: 16),
          label: Text(
              '${_pendingIssue!} (${AppLocalizations.of(context)!.pending})'),
          backgroundColor: Colors.orange.shade100,
        ),
      );
    }

    return SizedBox.shrink();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CompositedTransformTarget(
          link: _layerLink,
          child: TextFormField(
            controller: widget.controller,
            focusNode: _focusNode,
            decoration: InputDecoration(
              hintText: widget.hintText ??
                  AppLocalizations.of(context)!.entryDescription,
              border: OutlineInputBorder(),
              suffixIcon: _isResolving
                  ? Container(
                      width: 20,
                      height: 20,
                      padding: EdgeInsets.all(12),
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : null,
            ),
            maxLines: 3,
            minLines: 1,
          ),
        ),
        _buildIssueChip(),
      ],
    );
  }
}
