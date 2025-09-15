import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smooth_page_indicator/smooth_page_indicator.dart';
import 'package:timeasy/bloc/selected_project/selected_project_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';
import 'package:timeasy/components/project_header_component.dart';
import 'package:timeasy/models/project.dart';
import 'package:timeasy/tools/date_tools.dart';
import 'package:timeasy/views/statistics/weekly_statistics_widget.dart';

class WeeklyView extends StatelessWidget {
  final Project _project;

  WeeklyView(this._project);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: ProjectHeader(),
      body: BlocBuilder<SelectedProjectBloc, SelectedProjectState>(
        builder: (context, state) {
          Project projectToUse = _project;

          if (state is SelectedProjectSet && state.project != null) {
            projectToUse = state.project!;
          }

          return WeeklyViewWidget(projectToUse);
        },
      ),
    );
  }
}

class WeeklyViewWidget extends StatefulWidget {
  final Project _project;

  WeeklyViewWidget(this._project, {Key? key}) : super(key: key);

  @override
  _WeeklyViewState createState() {
    return new _WeeklyViewState(_project);
  }
}

class _WeeklyViewState extends State<WeeklyViewWidget> {
  final Project _project;
  int _calendarWeek = 0;
  int _year = 0;
  int _lastPosition = -1;
  final _dateTools = new DateTools();

  // Need to initialize the first page to such a high value to be able to swipe
  // backwards from the current week:
  final _pageController = new PageController(initialPage: 100000);

  _WeeklyViewState(this._project);

  @override
  void initState() {
    super.initState();
    _calendarWeek = _dateTools.getWeekNumber(DateTime.now()) -
        1; //need to subtract one because the page is flipped forward once on startup.
    _year = DateTime.now().year;
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          Expanded(
            child: PageView.builder(
              controller: _pageController,
              itemBuilder: (context, position) {
                if (position > _lastPosition) {
                  _calendarWeek++;
                  if (_calendarWeek > _dateTools.getNumberOfWeeks(_year)) {
                    _calendarWeek = 1;
                    _year++;
                  }
                } else if ((position < _lastPosition) && (_calendarWeek > 0)) {
                  _calendarWeek--;
                  if (_calendarWeek < 1) {
                    _year--;
                    _calendarWeek = _dateTools.getNumberOfWeeks(_year);
                  }
                }
                _lastPosition = position;
                return new WeeklyStatisticsWidget(_project, _calendarWeek, _year);
              },
            ),
          ),
          // SmoothPageIndicator that responds to swiping
          Padding(
            padding: const EdgeInsets.only(bottom: 16.0),
            child: SmoothPageIndicator(
              controller: _pageController,
              count: 5,
              effect: WormEffect(
                dotWidth: 10,
                dotHeight: 8,
                activeDotColor: Theme.of(context).colorScheme.primary,
                dotColor: Colors.grey.withOpacity(0.5),
                spacing: 8,
                radius: 4,
              ),
            ),
          ),
        ],
      ),
    );
  }

}
