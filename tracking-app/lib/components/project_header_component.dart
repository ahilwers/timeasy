import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_bloc.dart';
import 'package:timeasy/bloc/selected_project/selected_project_state.dart';
import 'package:timeasy/views/settings/settings_view.dart';

// Helper function to convert hex color string to Color
Color hexToColor(String hexString) {
  final buffer = StringBuffer();
  if (hexString.length == 6 || hexString.length == 7) buffer.write('ff');
  buffer.write(hexString.replaceFirst('#', ''));
  return Color(int.parse(buffer.toString(), radix: 16));
}

class ProjectHeader extends StatelessWidget implements PreferredSizeWidget {
  final String? title;
  final bool showBackButton;
  final VoidCallback? onBackPressed;
  final List<Widget>? actions;
  final bool showSettingsButton;

  const ProjectHeader({
    this.title,
    this.showBackButton = false,
    this.onBackPressed,
    this.actions,
    this.showSettingsButton = true,
  });

  @override
  Size get preferredSize => Size.fromHeight(kToolbarHeight);

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = isDark ? Colors.white : Colors.black;
    final backgroundColor = isDark ? Colors.black : Colors.white;
    
    return BlocBuilder<SelectedProjectBloc, SelectedProjectState>(
      builder: (context, state) {
        String displayTitle = title ?? '';

        if (title == null &&
            state is SelectedProjectSet &&
            state.project != null) {
          displayTitle = state.project!.name;
        }

        List<Widget> headerActions = [];

        // Only add settings button if showSettingsButton is true
        if (showSettingsButton) {
          headerActions.add(
            IconButton(
              icon: Icon(
                Icons.manage_accounts,
                color: textColor,
              ),
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => SettingsView()),
                );
              },
            ),
          );
        }

        if (actions != null) {
          headerActions.addAll(actions!);
        }

        return AppBar(
          backgroundColor: backgroundColor,
          elevation: 0,
          leading: showBackButton
              ? IconButton(
                  icon: Icon(
                    Icons.arrow_back,
                    color: textColor,
                  ),
                  onPressed: onBackPressed,
                )
              : null,
          title: Text(
            displayTitle,
            style: TextStyle(
              fontSize: 22,
              fontWeight: FontWeight.bold,
              color: textColor,
            ),
          ),
          centerTitle: true,
          actions: headerActions,
        );
      },
    );
  }
}
