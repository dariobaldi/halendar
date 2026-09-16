import 'package:flutter/material.dart';

import '../state/proposals_store.dart';
import 'history_screen.dart';
import 'proposals_list_screen.dart';
import 'settings_screen.dart';

/// Minimal bottom navigation for the three top-level destinations of the
/// MVP: the proposals that need action, the archive, and permissions.
class HomeShell extends StatefulWidget {
  const HomeShell({super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  final ProposalsStore _store = ProposalsStore();
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final screens = [
      ProposalsListScreen(store: _store),
      HistoryScreen(store: _store),
      const SettingsScreen(),
    ];

    return Scaffold(
      body: IndexedStack(index: _index, children: screens),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) => setState(() => _index = value),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.inbox_outlined),
            selectedIcon: Icon(Icons.inbox),
            label: 'Proposals',
          ),
          NavigationDestination(
            icon: Icon(Icons.history_outlined),
            selectedIcon: Icon(Icons.history),
            label: 'History',
          ),
          NavigationDestination(
            icon: Icon(Icons.settings_outlined),
            selectedIcon: Icon(Icons.settings),
            label: 'Settings',
          ),
        ],
      ),
    );
  }
}
