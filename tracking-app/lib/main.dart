import 'package:flex_color_scheme/flex_color_scheme.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:timeasy/bloc/authentication/authentication_bloc.dart';
import 'package:timeasy/bloc/authentication/authentication_event.dart';
import 'package:timeasy/bloc/authentication/authentication_state.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_bloc.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_event.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_state.dart';
import 'package:timeasy/bloc/selected_project/selected_project_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_event.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/components/project_swiper_component.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/services/background_sync_service.dart';
import 'package:timeasy/services/event_sync_service.dart';
import 'package:timeasy/services/internet_connection_service.dart';
import 'package:timeasy/views/project/project_list_view.dart';
import 'package:timeasy/views/statistics/weekly_view.dart';
import 'package:timeasy/views/theme.dart';
import 'package:timeasy/views/timeentry/time_entry_list_view.dart';

void main() {
  runApp(
    MultiBlocProvider(
      providers: [
        BlocProvider(create: (context) => AuthenticationBloc()),
        BlocProvider(create: (context) => InternetConnectionBloc()),
        BlocProvider(create: (context) => SynchronizationBloc()),
        BlocProvider(create: (context) => SelectedProjectBloc()),
        RepositoryProvider<BackgroundSyncService>(
          create: (context) {
            final syncBloc =
                BlocProvider.of<SynchronizationBloc>(context, listen: false);
            final authBloc =
                BlocProvider.of<AuthenticationBloc>(context, listen: false);
            return BackgroundSyncService("", syncBloc, authenticationBloc: authBloc);
          },
        ),
      ],
      child: MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  // Global navigator key for accessing context anywhere
  static final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();
  
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      navigatorKey: navigatorKey,
      title: 'timeasy',
      debugShowCheckedModeBanner: false,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('en', ''),
        Locale('de', ''),
      ],
      theme: getLightTheme(),
      darkTheme: getDarkTheme(),
      themeMode: ThemeMode.system,
      home: MainPage(),
    );
  }
}

class MainPage extends StatefulWidget {
  @override
  State<MainPage> createState() => _MainPageState();
}

class _MainPageState extends State<MainPage> with WidgetsBindingObserver {
  int _currentPageIndex = 0;
  final ProjectRepository _projectRepository = ProjectRepository();
  late InternetConnectionService _internetConnectionService;
  List<Project>? _projects;
  Project? _currentProject;
  bool _hasProjects = false;

  @override
  void initState() {
    super.initState();
    initializeDateFormatting();
    WidgetsBinding.instance.addObserver(this);
    
    // Initialize the EventSyncService with the available blocs
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        EventSyncService().initialize(
          authBloc: context.read<AuthenticationBloc>(),
          internetBloc: context.read<InternetConnectionBloc>(),
          syncBloc: context.read<SynchronizationBloc>(),
          initialToken: "",
        );
      }
    });
    
    _internetConnectionService = InternetConnectionService(
      onConnectionChanged: (bool hasInternet) {
        _setConnectionState(hasInternet);
        if (hasInternet) {
          context.read<AuthenticationBloc>().add(RefreshTokenEvent());
        }
      },
    );

    // Initialize the selected project
    _projectRepository.getLastUsedProjectOrDefault().then((Project? project) {
      if (project != null) {
        setState(() {
          _setCurrentProject(project);
          _hasProjects = true;
        });

        // Set the selected project in the SelectedProjectBloc
        context.read<SelectedProjectBloc>().add(
              SetSelectedProjectEvent(project),
            );
      } else {
        // No projects exist
        setState(() {
          _hasProjects = false;
          _currentPageIndex = 0; // Force to Home tab
        });

        // Clear the selected project in the SelectedProjectBloc
        context.read<SelectedProjectBloc>().add(
              ClearSelectedProjectEvent(),
            );
      }

      _loadProjects();
    });

    // Monitor changes to the selected project
    Future.delayed(Duration.zero, () {
      context.read<SelectedProjectBloc>().stream.listen((state) {
        if (state is SelectedProjectSet &&
            state.project != null) {
          setState(() {
            _setCurrentProject(state.project!);
            _hasProjects = true;
            
            // Reload projects to ensure we have the latest data
            _loadProjects();
          });
        } else if (state is SelectedProjectCleared) {
          setState(() {
            _hasProjects = false;
            _currentPageIndex = 0; // Force to Home tab
          });
        }
      });
    });
  }

  @override
  void dispose() {
    _internetConnectionService.dispose();
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      final connectionState = context.read<InternetConnectionBloc>().state;
      if (connectionState is InternetConnectionConnected) {
        context.read<AuthenticationBloc>().add(RefreshTokenEvent());
      }
    }
    super.didChangeAppLifecycleState(state);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: BlocListener<AuthenticationBloc, AuthenticationState>(
        listener: (context, state) {
          final backgroundSyncService = context.read<BackgroundSyncService>();
          final eventSyncService = EventSyncService();
          if (state is AuthenticationAuthenticated) {
            backgroundSyncService.updateToken(state.credentials.accessToken!);
            backgroundSyncService.startSync();
            eventSyncService.updateToken(state.credentials.accessToken!);
          } else {
            // Stop background sync on logout (AuthenticationInitial) or error
            backgroundSyncService.stopSync();
          }
        },
        child: Center(
          child: _getCurrentView(),
        ),
      ),
      bottomNavigationBar: _getNavigationBar(),
    );
  }

  Widget _getNavigationBar() {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final localizations = AppLocalizations.of(context)!;

    return BottomNavigationBar(
      currentIndex: _currentPageIndex,
      backgroundColor: isDark ? Colors.black : Colors.white,
      selectedItemColor: isDark ? Colors.white : Colors.black,
      unselectedItemColor: Colors.grey,
      onTap: (index) {
        // Only allow navigation to Home and Projects tabs if no projects exist
        if (!_hasProjects && index != 0 && index != 3) {
          return;
        }

        _loadProjects();
        setState(() {
          _currentPageIndex = index;
        });
      },
      items: [
        BottomNavigationBarItem(
          icon: Icon(Icons.home),
          label: localizations.navHome,
        ),
        BottomNavigationBarItem(
          icon: Icon(
            Icons.calendar_today,
            color: _hasProjects ? null : Colors.grey.shade300,
          ),
          label: localizations.navWeek,
        ),
        BottomNavigationBarItem(
          icon: Icon(
            Icons.access_time,
            color: _hasProjects ? null : Colors.grey.shade300,
          ),
          label: localizations.navTime,
        ),
        BottomNavigationBarItem(
          icon: Icon(Icons.list),
          label: localizations.navProjects,
        ),
      ],
    );
  }

  Widget _getCurrentView() {
    // If no projects exist and we're not on the Home or Projects tab,
    // force to the Home tab
    if (!_hasProjects && _currentPageIndex != 0 && _currentPageIndex != 3) {
      setState(() {
        _currentPageIndex = 0;
      });
    }

    switch (_currentPageIndex) {
      case 0:
        return _playButtonView();
      case 1:
        return _hasProjects ? WeeklyView(_currentProject!) : _playButtonView();
      case 2:
        return _hasProjects
            ? TimeEntryListView(_currentProject!)
            : _playButtonView();
      case 3:
        return ProjectListView();
      default:
        return _playButtonView();
    }
  }

  Widget _playButtonView() {
    return ProjectSwiper();
  }

  _loadProjects() {
    _projectRepository.getAllProjects().then((List<Project> projectsFromDb) {
      setState(() {
        _projects = projectsFromDb;
        _hasProjects = projectsFromDb.isNotEmpty;
        
        // If we have projects but no current project is set, set the first one
        if (_hasProjects && _currentProject == null && projectsFromDb.isNotEmpty) {
          _setCurrentProject(projectsFromDb[0]);
          
          // Update the selected project in the bloc
          context.read<SelectedProjectBloc>().add(
                SetSelectedProjectEvent(projectsFromDb[0]),
              );
        }
      });
    });
  }

  _setCurrentProject(Project project) {
    _currentProject = project;
    _projectRepository.saveLastUsedProject(project);
  }

  _setConnectionState(bool hasInternet) {
    if (hasInternet) {
      context
          .read<InternetConnectionBloc>()
          .add(InternetConnectionConnectedEvent());
    } else {
      context
          .read<InternetConnectionBloc>()
          .add(InternetConnectionDisconnectedEvent());
    }
  }
}
