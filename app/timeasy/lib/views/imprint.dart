import 'package:flutter/material.dart';

import 'package:flutter_gen/gen_l10n/app_localizations.dart';

class Imprint extends StatelessWidget {
  const Imprint({Key? key}) : super(key: key);

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

  bool _isDarkMode(BuildContext context) {
    var brightness = MediaQuery.of(context).platformBrightness;
    return brightness == Brightness.dark;
  }
}
