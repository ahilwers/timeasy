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
import 'package:timeasy/bloc/synchronization/synchronization_bloc.dart';
import 'package:timeasy/components/project_swiper_component.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/models/time_entry.dart';
import 'package:timeasy/repositories/project_repository.dart';
import 'package:timeasy/repositories/time_entry_repository.dart';
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
      localizationsDelegates: [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: [
        Locale('en', ''),
        Locale('de', ''),
      ],
      theme: FlexColorScheme.light(colors: timeasyTheme.light).toTheme,
      darkTheme: FlexColorScheme.dark(colors: timeasyTheme.dark).toTheme,
      // Use dark or light theme based on system setting.
      themeMode: ThemeMode.system,
      home: MainPage(title: 'timeasy'),
    );
  }
}

class MainPage extends StatefulWidget {
  final String? title;

  MainPage({Key? key, this.title}) : super(key: key);

  @override
  _MainPageState createState() {
    return new _MainPageState();
  }
}

enum AppState { RUNNING, STOPPED }

class _MainPageState extends State<MainPage>
    with WidgetsBindingObserver, SingleTickerProviderStateMixin {
  AppState _currentState = AppState.STOPPED;
  late Project _currentProject;
  List<Project>? _projects;
  int _currentPageIndex = 0;

  final ProjectRepository _projectRepository = new ProjectRepository();
  final TimeEntryRepository _timeEntryRepository = new TimeEntryRepository();
  late AnimationController buttonAnimationController;
  late InternetConnectionService _internetConnectionService;

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
    buttonAnimationController = new AnimationController(
      vsync: this,
      duration: Duration(milliseconds: 1000),
    );
    var projectRepository = new ProjectRepository();
    projectRepository
        .getLastUsedProjectOrDefault("Project 1")
        .then((Project project) {
      setState(() {
        _setCurrentProject(project);
      });
      _loadProjects();
      _updateAppState();
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

  void _setAppState(AppState state) {
    switch (state) {
      case AppState.RUNNING:
        if (_currentState == AppState.STOPPED) {
          buttonAnimationController.forward();
        }
        break;
      case AppState.STOPPED:
        if (_currentState == AppState.RUNNING) {
          buttonAnimationController.reverse();
        }
        break;
    }
    setState(() {
      _currentState = state;
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

  void _startTiming() async {
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry(_currentProject.id);
    await repository.getLatestOpenTimeEntryOrCreateNew(_currentProject.id);
    _setAppState(AppState.RUNNING);
  }

  void _stopTiming() async {
    var repository = new TimeEntryRepository();
    await repository.closeLatestTimeEntry(_currentProject.id);
    _setAppState(AppState.STOPPED);
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
    // return Scaffold(
    //   appBar: AppBar(
    //     title: Text('timeasy'),
    //     backgroundColor: Theme.of(context).primaryColor,
    //     automaticallyImplyLeading: false,
    //     actions: [
    //       IconButton(
    //         icon: Icon(Icons.manage_accounts),
    //         onPressed: () {
    //           Navigator.push(
    //             context,
    //             MaterialPageRoute(builder: (context) => SettingsView()),
    //           );
    //         },
    //       ),
    //     ],
    //   ),
    //   body: Column(
    //     mainAxisAlignment: MainAxisAlignment.center,
    //     children: <Widget>[
    //       Align(
    //         alignment: Alignment.center,
    //         child: new RawMaterialButton(
    //           onPressed: _toggleState,
    //           child: new AnimatedIcon(
    //             icon: AnimatedIcons.play_pause,
    //             color: Colors.white,
    //             size: 128.0,
    //             progress: buttonAnimationController,
    //           ),
    //           shape: new CircleBorder(),
    //           elevation: 2.0,
    //           fillColor: Theme.of(context).primaryColor,
    //           padding: const EdgeInsets.all(15.0),
    //         ),
    //       ),
    //       _projects == null
    //           ? Text(AppLocalizations.of(context)!.loadingProject)
    //           : BlocListener<SynchronizationBloc, SynchronizationState>(
    //               listener: (context, state) {
    //                 if (state is SynchronizationSuccess) {
    //                   _loadProjects();
    //                   _updateAppState();
    //                 }
    //               },
    //               child: new DropdownButton<String>(
    //                 value: _currentProject.id,
    //                 items: _projects!.map(
    //                   (Project value) {
    //                     return new DropdownMenuItem<String>(
    //                       value: value.id,
    //                       child: new Text(value.name),
    //                     );
    //                   },
    //                 ).toList(),
    //                 onChanged: (String? value) {
    //                   _projectRepository.getProjectById(value!).then(
    //                     (Project? projectFromDb) {
    //                       setState(
    //                         () {
    //                           _setCurrentProject(projectFromDb!);
    //                         },
    //                       );
    //                       _updateAppState();
    //                     },
    //                   );
    //                 },
    //               ),
    //             ),
    //     ],
    //   ),
    // );
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

  _updateAppState() {
    // Set the current state if there's a timing already running:
    _timeEntryRepository
        .getLatestOpenTimeEntry(_currentProject.id)
        .then((TimeEntry? entry) {
      if (entry != null) {
        _setAppState(AppState.RUNNING);
      } else {
        _setAppState(AppState.STOPPED);
      }
    });
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
