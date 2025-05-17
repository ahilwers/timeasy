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
            return BackgroundSyncService("", syncBloc);
          },
        ),
      ],
      child: MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
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
      theme: FlexColorScheme.light(colors: timeasyTheme.light).toTheme,
      darkTheme: FlexColorScheme.dark(colors: timeasyTheme.dark).toTheme,
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
  late Project _currentProject;

  @override
  void initState() {
    super.initState();
    initializeDateFormatting();
    WidgetsBinding.instance.addObserver(this);
    _internetConnectionService = InternetConnectionService(
      onConnectionChanged: (bool hasInternet) {
        _setConnectionState(hasInternet);
        if (hasInternet) {
          context.read<AuthenticationBloc>().add(RefreshTokenEvent());
        }
      },
    );
    
    // Initialize the selected project
    _projectRepository
        .getLastUsedProjectOrDefault("Project 1")
        .then((Project project) {
      setState(() {
        _setCurrentProject(project);
      });
      
      // Set the selected project in the SelectedProjectBloc
      context.read<SelectedProjectBloc>().add(
            SetSelectedProjectEvent(project),
          );
      
      _loadProjects();
    });

    // Monitor changes to the selected project
    Future.delayed(Duration.zero, () {
      context.read<SelectedProjectBloc>().stream.listen((state) {
        if (state is SelectedProjectSet &&
            state.project != null &&
            _currentProject != null &&
            state.project!.id != _currentProject.id) {
          setState(() {
            _setCurrentProject(state.project!);
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
          if (state is AuthenticationAuthenticated) {
            backgroundSyncService.updateToken(state.credentials.accessToken!);
            backgroundSyncService.startSync();
          } else if (state is AuthenticationError) {
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

    return BottomNavigationBar(
      currentIndex: _currentPageIndex,
      backgroundColor: isDark ? Colors.black : Colors.white,
      selectedItemColor: isDark ? Colors.white : Colors.black,
      unselectedItemColor: Colors.grey,
      onTap: (index) {
        _loadProjects();
        setState(() {
          _currentPageIndex = index;
        });
      },
      items: const [
        BottomNavigationBarItem(
          icon: Icon(Icons.home),
          label: 'Home',
        ),
        BottomNavigationBarItem(
          icon: Icon(Icons.calendar_today),
          label: 'Week',
        ),
        BottomNavigationBarItem(
          icon: Icon(Icons.access_time),
          label: 'Time',
        ),
        BottomNavigationBarItem(
          icon: Icon(Icons.list),
          label: 'Projects',
        ),
      ],
    );
  }

  Widget _getCurrentView() {
    switch (_currentPageIndex) {
      case 0:
        return _playButtonView();
      case 1:
        return WeeklyView(_currentProject);
      case 2:
        return TimeEntryListView(_currentProject);
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
      });
    });
  }

  _setCurrentProject(Project project) {
    _currentProject = project;
    _projectRepository.saveLastUsedProject(project);
  }

  void _setConnectionState(bool hasInternet) {
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
