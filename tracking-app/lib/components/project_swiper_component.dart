import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:smooth_page_indicator/smooth_page_indicator.dart';
import 'package:timeasy/bloc/selected_project/selected_project_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_event.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/bloc/synchronization/synchronization_state.dart';
import 'package:timeasy/components/project_header_component.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/models/time_entry.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
import 'package:timeasy/views/project/project_edit_view.dart';
import 'dart:async';

// Helper function to convert hex color string to Color
Color hexToColor(String hexString) {
  final buffer = StringBuffer();
  if (hexString.length == 6 || hexString.length == 7) buffer.write('ff');
  buffer.write(hexString.replaceFirst('#', ''));
  return Color(int.parse(buffer.toString(), radix: 16));
}

// AppState Enum for the timing status
enum AppState { RUNNING, STOPPED }

class ProjectSwiper extends StatefulWidget {
  const ProjectSwiper();

  @override
  State<ProjectSwiper> createState() => _ProjectSwiperState();
}

class _ProjectSwiperState extends State<ProjectSwiper>
    with TickerProviderStateMixin {
  final ProjectRepository _projectRepository = new ProjectRepository();
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  final TextEditingController _descriptionController = TextEditingController();
  List<Project> _projects = [];
  List<String> _suggestions = [];
  bool _isLoading = true;
  int _currentPage = 0;
  late PageController _controller;
  late AnimationController _buttonAnimationController;
  AppState _currentState = AppState.STOPPED;
  Timer? _debounceTimer;
  TimeEntry? _currentOpenTimeEntry;

  @override
  void initState() {
    super.initState();
    _controller = PageController();
    _buttonAnimationController = AnimationController(
      vsync: this,
      duration: Duration(milliseconds: 1000),
    );
    _descriptionController.addListener(_onDescriptionChanged);
    _loadProjects();
  }

  @override
  void dispose() {
    _buttonAnimationController.dispose();
    _controller.dispose();
    _descriptionController.removeListener(_onDescriptionChanged);
    _descriptionController.dispose();
    _debounceTimer?.cancel();
    super.dispose();
  }

  // This method is called whenever the text changes
  void _onDescriptionChanged() {
    if (_currentState == AppState.RUNNING && _currentOpenTimeEntry != null) {
      // Cancel the previous timer if it exists
      _debounceTimer?.cancel();
      
      // Start a new timer
      _debounceTimer = Timer(Duration(milliseconds: 500), () {
        _saveDescription();
      });
    }
  }

  // Save the description to the current open time entry
  Future<void> _saveDescription() async {
    if (_currentOpenTimeEntry != null) {
      final description = _descriptionController.text.trim();
      
      // Only save if the description has changed
      if (_currentOpenTimeEntry!.description != description) {
        _currentOpenTimeEntry!.description = description;
        await _timeEntryRepository.updateTimeEntry(_currentOpenTimeEntry!);
      }
    }
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // When the widget is rebuilt (e.g., when switching tabs), update the page
    if (!_isLoading && _projects.isNotEmpty) {
      final selectedProjectState = context.read<SelectedProjectBloc>().state;
      if (selectedProjectState is SelectedProjectSet && selectedProjectState.project != null) {
        final selectedProject = selectedProjectState.project!;
        final index = _projects.indexWhere((p) => p.id == selectedProject.id);
        if (index != -1 && index != _currentPage) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (mounted) {
              _controller.jumpToPage(index);
              setState(() {
                _currentPage = index;
              });
              _checkTimingStatus();
            }
          });
        }
      }
    }
  }

  Future<void> _loadProjects() async {
    setState(() {
      _isLoading = true;
    });

    var projects = await _projectRepository.getAllProjects();

    setState(() {
      _projects = projects;
      _isLoading = false;

      // If there are no projects, set the current page to the "add new project" page
      if (_projects.isEmpty) {
        // We need to use a small delay to ensure the PageView is built
        Future.delayed(Duration(milliseconds: 100), () {
          if (_controller.hasClients) {
            _controller.jumpToPage(0); // The "add new project" page will be at index 0
          }
        });
      } else {
        // Find the index of the current project
        final currentProject = context.read<SelectedProjectBloc>().state.project;
        if (currentProject != null) {
          final index = _projects.indexWhere((p) => p.id == currentProject.id);
          if (index != -1) {
            _currentPage = index;
            // We need to use a small delay to ensure the PageView is built
            Future.delayed(Duration(milliseconds: 100), () {
              if (_controller.hasClients) {
                _controller.jumpToPage(index);
              }
            });
          }
        }
      }
    });

    // Load suggestions for the current project if available
    if (_projects.isNotEmpty && _currentPage < _projects.length) {
      _loadSuggestions(_projects[_currentPage].id);
    }

    _checkTimingStatus();
  }

  void _loadSuggestions(String projectId) async {
    final suggestions = await _timeEntryRepository.getLastUniqueDescriptions(projectId);
    setState(() {
      _suggestions = suggestions;
    });
  }

  void _checkTimingStatus() {
    if (_projects.isEmpty || _currentPage >= _projects.length) {
      return;
    }

    final currentProject = _projects[_currentPage];
    _timeEntryRepository
        .getLatestOpenTimeEntry(currentProject.id)
        .then((TimeEntry? entry) {
      if (mounted) {
        if (entry != null) {
          setState(() {
            _currentState = AppState.RUNNING;
            _buttonAnimationController.forward();
            _currentOpenTimeEntry = entry;
            
            // Update the description field with the current time entry's description
            if (entry.description != null && entry.description!.isNotEmpty) {
              _descriptionController.text = entry.description!;
            } else {
              _descriptionController.clear();
            }
          });
        } else {
          setState(() {
            _currentState = AppState.STOPPED;
            _buttonAnimationController.reverse();
            _currentOpenTimeEntry = null;
            _descriptionController.clear();
          });
        }
      }
    });
  }

  void _startTiming() async {
    if (_projects.isEmpty || _currentPage >= _projects.length) {
      return;
    }

    final currentProject = _projects[_currentPage];
    await _timeEntryRepository.closeLatestTimeEntry(currentProject.id);

    // Create a new time entry
    final timeEntry = TimeEntry(currentProject.id);

    // Set description if provided
    final description = _descriptionController.text.trim();
    if (description.isNotEmpty) {
      timeEntry.description = description;
    }

    // Add the time entry
    await _timeEntryRepository.addTimeEntry(timeEntry);
    
    // Store reference to the current open time entry
    _currentOpenTimeEntry = timeEntry;

    setState(() {
      _currentState = AppState.RUNNING;
      _buttonAnimationController.forward();
      // Don't clear the description field anymore
      // _descriptionController.clear();
    });
  }

  void _stopTiming() async {
    if (_projects.isEmpty || _currentPage >= _projects.length) {
      return;
    }

    final currentProject = _projects[_currentPage];
    await _timeEntryRepository.closeLatestTimeEntry(currentProject.id);

    setState(() {
      _currentState = AppState.STOPPED;
      _buttonAnimationController.reverse();
    });
  }

  void _toggleState() {
    switch (_currentState) {
      case AppState.STOPPED:
        _startTiming();
        break;
      case AppState.RUNNING:
        _stopTiming();
        break;
    }
  }

  Future<void> _createNewProject() async {
    // Navigate to the project edit view
    final result = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ProjectEditView(),
        fullscreenDialog: true,
      ),
    );

    // Check if a project was returned (user clicked Save)
    if (result != null && result is Project) {
      // Reload projects
      await _loadProjects();
      
      // Explicitly set the newly created project in the SelectedProjectBloc
      context.read<SelectedProjectBloc>().add(
        SetSelectedProjectEvent(result),
      );
      
      // Find the index of the new project and jump to it
      final index = _projects.indexWhere((p) => p.id == result.id);
      if (index != -1) {
        setState(() {
          _currentPage = index;
        });
        
        if (_controller.hasClients) {
          _controller.jumpToPage(index);
        }
      }
    } else {
      // User canceled, just reload projects
      _loadProjects();
    }
  }

  @override
  Widget build(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    if (_isLoading) {
      return Center(child: CircularProgressIndicator());
    }

    // Get the current project if available
    final project = _projects.isNotEmpty && _currentPage < _projects.length
        ? _projects[_currentPage]
        : null;

    // Get project color for UI elements
    final projectColor = project != null
        ? hexToColor(project.color)
        : Colors.grey;

    return Scaffold(
      backgroundColor: isDark ? Colors.black : Colors.white,
      body: SafeArea(
        child: BlocListener<SynchronizationBloc, SynchronizationState>(
          listener: (context, state) {
            if (state is SynchronizationSuccess) {
              _loadProjects();
            }
          },
          child: Column(
            children: [
              ProjectHeader(),
              Expanded(
                child: PageView.builder(
                  controller: _controller,
                  itemCount: _projects.isEmpty ? 1 : _projects.length + 1, // If no projects, just show the add page
                  onPageChanged: (int page) {
                    setState(() {
                      _currentPage = page;
                    });

                    // Update the selected project in the bloc only if we're on a valid project
                    if (_projects.isNotEmpty && page < _projects.length) {
                      context.read<SelectedProjectBloc>().add(
                            SetSelectedProjectEvent(_projects[page]),
                          );

                      // Load suggestions for the new project
                      _loadSuggestions(_projects[page].id);

                      // Check timing status for the new project
                      _checkTimingStatus();
                    }
                  },
                  itemBuilder: (context, index) {
                    // If this is the last page (add new project page) or there are no projects
                    if (_projects.isEmpty || index == _projects.length) {
                      return Center(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            GestureDetector(
                              onTap: _createNewProject,
                              child: Stack(
                                alignment: Alignment.center,
                                children: [
                                  Container(
                                    width: 180,
                                    height: 180,
                                    decoration: BoxDecoration(
                                      shape: BoxShape.circle,
                                      border: Border.all(
                                        color: Colors.grey,
                                        width: 12,
                                      ),
                                    ),
                                  ),
                                  Container(
                                    width: 130,
                                    height: 130,
                                    decoration: BoxDecoration(
                                      shape: BoxShape.circle,
                                      color: Colors.grey,
                                    ),
                                    child: Center(
                                      child: Icon(
                                        Icons.add,
                                        color: Colors.white,
                                        size: 50,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 30),
                            Padding(
                              padding: const EdgeInsets.symmetric(horizontal: 40),
                              child: Text(
                                localizations.addNewProject,
                                style: TextStyle(
                                  fontSize: 18,
                                  color: isDark ? Colors.white : Colors.black,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                          ],
                        ),
                      );
                    }

                    // Regular project page
                    final proj = _projects[index];
                    return Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          GestureDetector(
                            onTap: index == _currentPage ? _toggleState : null,
                            child: Stack(
                              alignment: Alignment.center,
                              children: [
                                Container(
                                  width: 180,
                                  height: 180,
                                  decoration: BoxDecoration(
                                    shape: BoxShape.circle,
                                    border: Border.all(
                                      color: hexToColor(proj.color),
                                      width: 12,
                                    ),
                                  ),
                                ),
                                Container(
                                  width: 130,
                                  height: 130,
                                  decoration: BoxDecoration(
                                    shape: BoxShape.circle,
                                    color: hexToColor(proj.color),
                                  ),
                                  child: Center(
                                    child: index == _currentPage
                                        ? AnimatedIcon(
                                            icon: AnimatedIcons.play_pause,
                                            progress: _buttonAnimationController,
                                            color: Colors.white,
                                            size: 50,
                                          )
                                        : Icon(
                                            Icons.play_arrow,
                                            color: Colors.white,
                                            size: 50,
                                          ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 30),
                          Padding(
                            padding: const EdgeInsets.symmetric(horizontal: 40),
                            child: Autocomplete<String>(
                              optionsBuilder: (TextEditingValue textEditingValue) {
                                if (textEditingValue.text.isEmpty) {
                                  return const Iterable<String>.empty();
                                }
                                return _suggestions.where((String option) {
                                  return option.toLowerCase()
                                      .contains(textEditingValue.text.toLowerCase());
                                });
                              },
                              onSelected: (String selection) {
                                _descriptionController.text = selection;
                              },
                              fieldViewBuilder: (
                                BuildContext context,
                                TextEditingController fieldController,
                                FocusNode fieldFocusNode,
                                VoidCallback onFieldSubmitted,
                              ) {
                                // Sync the autocomplete controller with our description controller
                                fieldController.text = _descriptionController.text;
                                fieldController.addListener(() {
                                  _descriptionController.text = fieldController.text;
                                });
                                
                                return TextField(
                                  controller: fieldController,
                                  focusNode: fieldFocusNode,
                                  style: TextStyle(
                                    color: isDark ? Colors.white : Colors.black,
                                  ),
                                  decoration: InputDecoration(
                                    hintText: localizations.whatAreYouDoingNow,
                                    hintStyle: TextStyle(
                                      color: isDark ? Colors.white70 : Colors.black54,
                                    ),
                                    filled: true,
                                    fillColor:
                                        isDark ? Colors.white12 : Colors.black12,
                                    border: OutlineInputBorder(
                                      borderRadius: BorderRadius.circular(12),
                                      borderSide: BorderSide.none,
                                    ),
                                  ),
                                );
                              },
                              optionsViewBuilder: (
                                BuildContext context,
                                AutocompleteOnSelected<String> onSelected,
                                Iterable<String> options,
                              ) {
                                return Align(
                                  alignment: Alignment.topLeft,
                                  child: Material(
                                    elevation: 4.0,
                                    color: isDark ? Colors.grey[800] : Colors.white,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                    child: Container(
                                      width: 300,
                                      constraints: BoxConstraints(maxHeight: 200),
                                      child: ListView.builder(
                                        padding: EdgeInsets.zero,
                                        shrinkWrap: true,
                                        itemCount: options.length,
                                        itemBuilder: (BuildContext context, int index) {
                                          final String option = options.elementAt(index);
                                          return InkWell(
                                            onTap: () {
                                              onSelected(option);
                                            },
                                            child: ListTile(
                                              title: Text(
                                                option,
                                                style: TextStyle(
                                                  color: isDark ? Colors.white : Colors.black,
                                                ),
                                              ),
                                            ),
                                          );
                                        },
                                      ),
                                    ),
                                  ),
                                );
                              },
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ),
              const SizedBox(height: 16),
              SmoothPageIndicator(
                controller: _controller,
                count: _projects.isEmpty ? 1 : _projects.length + 1, // If no projects, just show the add page
                effect: WormEffect(
                  dotHeight: 8,
                  dotWidth: 10,
                  activeDotColor: _currentPage < _projects.length
                      ? hexToColor(_projects[_currentPage].color)
                      : Colors.grey,
                  dotColor: Colors.grey.shade300,
                  spacing: 8,
                  radius: 4,
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
  }
}
