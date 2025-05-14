import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smooth_page_indicator/smooth_page_indicator.dart';
import 'package:timeasy/bloc/selected_project/selected_project_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_event.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/repositories/project_repository.dart';

// Helper function to convert hex color string to Color
Color hexToColor(String hexString) {
  final buffer = StringBuffer();
  if (hexString.length == 6 || hexString.length == 7) buffer.write('ff');
  buffer.write(hexString.replaceFirst('#', ''));
  return Color(int.parse(buffer.toString(), radix: 16));
}

class ProjectSwiper extends StatefulWidget {
  const ProjectSwiper();

  @override
  State<ProjectSwiper> createState() => _ProjectSwiperState();
}

class _ProjectSwiperState extends State<ProjectSwiper> {
  final PageController _controller = PageController();
  List<Project>? _projects;
  final ProjectRepository _projectRepository = new ProjectRepository();

  int _currentPage = 0;
  bool _initialScrollDone = false;

  @override
  void initState() {
    super.initState();
    _loadProjects();
  }

  @override
  Widget build(BuildContext context) {
    final project = _projects == null ? Project() : _projects![_currentPage];
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final isAddButton = false; //Todo: need to implement
    final projectColor = hexToColor(project.color);

    // Check if we need to scroll to the selected project
    if (_projects != null && _projects!.isNotEmpty && !_initialScrollDone) {
      _scrollToSelectedProject(context);
    }

    // When projects are loaded, set the current project in the SelectedProjectBloc
    if (_projects != null && _projects!.isNotEmpty) {
      context.read<SelectedProjectBloc>().add(
            SetSelectedProjectEvent(_projects![_currentPage]),
          );
    }

    return Scaffold(
      backgroundColor: isDark ? Colors.black : Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: Center(
                child: Text(
                  _projects != null && _projects!.isNotEmpty ? _projects![_currentPage].name : '',
                  style: TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.bold,
                    color: isDark ? Colors.white : Colors.black,
                  ),
                ),
              ),
            ),
            Expanded(
              child: PageView.builder(
                controller: _controller,
                onPageChanged: (index) {
                  setState(() => _currentPage = index);
                  
                  // Update the selected project when page changes
                  if (_projects != null && _projects!.isNotEmpty) {
                    context.read<SelectedProjectBloc>().add(
                          SetSelectedProjectEvent(_projects![index]),
                        );
                  }
                },
                itemCount: _projects == null ? 0 : _projects!.length,
                itemBuilder: (context, index) {
                  final proj = _projects![index];
                  return Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Stack(
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
                                child: Icon(
                                  isAddButton ? Icons.add : Icons.play_arrow,
                                  color: Colors.white,
                                  size: 50,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 30),
                        Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 40),
                          child: isAddButton
                              ? const SizedBox.shrink()
                              : TextField(
                                  style: TextStyle(
                                    color: isDark ? Colors.white : Colors.black,
                                  ),
                                  decoration: InputDecoration(
                                    hintText: "Was machst du gerade?",
                                    hintStyle: TextStyle(
                                      color: isDark
                                          ? Colors.white70
                                          : Colors.black54,
                                    ),
                                    filled: true,
                                    fillColor: isDark
                                        ? Colors.white12
                                        : Colors.black12,
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
              count: _projects == null ? 0 : _projects!.length,
              effect: WormEffect(
                activeDotColor: projectColor,
                dotColor: Colors.grey,
                dotHeight: 8,
                dotWidth: 8,
              ),
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  _loadProjects() {
    _projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      setState(() {
        _projects = projectsFromDb;
      });
    });
  }
  
  // Scroll to the selected project without triggering the event
  void _scrollToSelectedProject(BuildContext context) {
    final selectedProjectState = context.read<SelectedProjectBloc>().state;
    
    if (selectedProjectState is SelectedProjectSet && 
        selectedProjectState.project != null && 
        _projects != null) {
      
      // Find the index of the selected project
      final selectedProjectId = selectedProjectState.project!.id;
      final selectedIndex = _projects!.indexWhere((p) => p.id == selectedProjectId);
      
      if (selectedIndex != -1 && selectedIndex != _currentPage) {
        // Jump to the page without animation to avoid triggering onPageChanged
        _controller.jumpToPage(selectedIndex);
        setState(() {
          _currentPage = selectedIndex;
          _initialScrollDone = true;
        });
      } else {
        _initialScrollDone = true;
      }
    } else {
      _initialScrollDone = true;
    }
  }
}
