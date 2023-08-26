import 'package:flutter/material.dart';
import 'package:package_info_plus/package_info_plus.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class Imprint extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(body: new ImprintWidget());
  }
}

class ImprintWidget extends StatefulWidget {
  @override
  _ImprintWidgetState createState() {
    return new _ImprintWidgetState();
  }
}

class _ImprintWidgetState extends State<ImprintWidget> {
  _VersionInfo _versionInfo = new _VersionInfo();

  @override
  initState() {
    super.initState();
    _loadVersionInfo();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(AppLocalizations.of(context)!.info),
        backgroundColor: Theme.of(context).primaryColor,
      ),
      body: Container(
        margin: EdgeInsets.all(10),
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: _isDarkMode(context) ? Image.asset('assets/hourglass_gradient_lightgrey.png') : Image.asset('assets/hourglass_gradient_black.png'),
              ),
              Text('\nVersion: ${_versionInfo.version}, Build: ${_versionInfo.buildNumber}\n',
                  textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('\n(c) 2021 Achim Hilwers Software\n', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('\nAchim Hilwers Software\n', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
              Text('Schützenstraße 17', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('26676 Barßel\n', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('www.timeasy.de', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('info@timeasy.de', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
            ],
          ),
        ),
      ),
    );
  }

  void _loadVersionInfo() {
    PackageInfo.fromPlatform().then((packageInfo) => {
          setState(() {
            _versionInfo.appName = packageInfo.appName;
            _versionInfo.packageName = packageInfo.packageName;
            _versionInfo.version = packageInfo.version;
            _versionInfo.buildNumber = packageInfo.buildNumber;
          })
        });
  }

  bool _isDarkMode(BuildContext context) {
    var brightness = MediaQuery.of(context).platformBrightness;
    return brightness == Brightness.dark;
  }
}

class _VersionInfo {
  String appName = '';
  String packageName = '';
  String version = '';
  String buildNumber = '';
}
