import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_event.dart';
import 'package:timeasy/bloc/internetconnection/internet_connection_state.dart';

class InternetConnectionBloc
    extends Bloc<InternetConnectionEvent, InternetConnectionState> {
  InternetConnectionBloc() : super(InternetConnectionDisconnected()) {
    on<InternetConnectionConnectedEvent>((event, emit) {
      emit(InternetConnectionConnected());
    });
    on<InternetConnectionDisconnectedEvent>((event, emit) {
      emit(InternetConnectionDisconnected());
    });
  }
}
