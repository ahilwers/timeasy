import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
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
  final ProjectRepository _projectRepository = ProjectRepository();
  final TimeEntryRepository _timeEntryRepository = TimeEntryRepository();
  late PageController _controller;
  late AnimationController _buttonAnimationController;
  List<Project> _projects = [];
  int _currentPage = 0;
  bool _isLoading = true;
  AppState _currentState = AppState.STOPPED;
  TextEditingController _descriptionController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _controller = PageController(initialPage: 0);
    _buttonAnimationController = AnimationController(
      vsync: this,
      duration: Duration(milliseconds: 1000),
    );
    
    // Load projects immediately after initialization
    _loadProjects();
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

  @override
  void dispose() {
    _buttonAnimationController.dispose();
    _controller.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  void _loadProjects() {
    setState(() {
      _isLoading = true;
    });

    _projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      if (mounted) {
        if (projectsFromDb.isEmpty) {
          setState(() {
            _isLoading = false;
          });
          return;
        }
        
        setState(() {
          _projects = projectsFromDb;
          _isLoading = false;
        });
        
        // After loading projects, check if there's a selected project to scroll to
        final selectedProjectState = context.read<SelectedProjectBloc>().state;
        if (selectedProjectState is SelectedProjectSet && selectedProjectState.project != null) {
          final selectedProject = selectedProjectState.project!;
          final index = _projects.indexWhere((p) => p.id == selectedProject.id);
          if (index != -1) {
            WidgetsBinding.instance.addPostFrameCallback((_) {
              if (mounted) {
                _controller.jumpToPage(index);
                setState(() {
                  _currentPage = index;
                });
                
                // Ensure the current project is set in the SelectedProjectBloc
                context.read<SelectedProjectBloc>().add(
                      SetSelectedProjectEvent(_projects[_currentPage]),
                    );
              }
            });
          }
        } else if (_projects.isNotEmpty) {
          // If no project is selected but we have projects, select the first one
          context.read<SelectedProjectBloc>().add(
                SetSelectedProjectEvent(_projects[0]),
              );
        }
        
        // Check the timing status for the current project
        _checkTimingStatus();
      }
    }).catchError((error) {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
        print("Error loading projects: $error");
      }
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
          });
        } else {
          setState(() {
            _currentState = AppState.STOPPED;
            _buttonAnimationController.reverse();
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
    
    setState(() {
      _currentState = AppState.RUNNING;
      _buttonAnimationController.forward();
      _descriptionController.clear();
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

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return Center(child: CircularProgressIndicator());
    }

    if (_projects.isEmpty) {
      return Center(child: Text('No projects found'));
    }

    final project = _projects[_currentPage];
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final projectColor = hexToColor(project.color);

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
                  itemCount: _projects.length,
                  onPageChanged: (int page) {
                    setState(() {
                      _currentPage = page;
                    });
                    
                    // Update the selected project in the bloc
                    context.read<SelectedProjectBloc>().add(
                          SetSelectedProjectEvent(_projects[page]),
                        );
                    
                    // Check timing status for the new project
                    _checkTimingStatus();
                  },
                  itemBuilder: (context, index) {
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
                            child: TextField(
                              controller: _descriptionController,
                              style: TextStyle(
                                color: isDark ? Colors.white : Colors.black,
                              ),
                              decoration: InputDecoration(
                                hintText: "What are you doing right now?",
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
                count: _projects.length,
                effect: ExpandingDotsEffect(
                  dotHeight: 8,
                  dotWidth: 8,
                  activeDotColor: projectColor,
                  dotColor: Colors.grey.shade300,
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
