import 'package:flutter/material.dart';

class Imprint extends StatelessWidget {
  const Imprint({Key key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
        appBar: AppBar(
          title: Text("Impressum"),
        ),
        body: Container(
          margin: EdgeInsets.all(10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('\nAchim Hilwers Software\n', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
              Text('Schützenstraße 17', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('26676 Barßel\n', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('www.timeasy.de', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
              Text('info@timeasy.de', textAlign: TextAlign.left, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
            ],
          ),
        ));
  }
}
