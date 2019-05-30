import 'package:flutter/material.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'timeasy',
      home: Scaffold(
        appBar: AppBar(
          title: Text('timeasy'),
        ),
        body: Center(
          child : new RawMaterialButton(
            onPressed: () {},
            child: new Icon(
              Icons.play_arrow,
              color: Colors.blue,
              size: 128.0,
            ),
            shape: new CircleBorder(),
            elevation: 2.0,
            fillColor: Colors.white,
            padding: const EdgeInsets.all(15.0),
          ),
        ),
      ),
    );
  }
}
