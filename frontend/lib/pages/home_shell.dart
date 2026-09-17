import 'package:flutter/material.dart';

import '../services/auth.dart';
import '../services/notifications.dart';
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
  void initState() {
    super.initState();
    _store.init();
  }

  @override
  void dispose() {
    _store.end();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final screens = [
      ProposalsListScreen(store: _store),
      HistoryScreen(store: _store),
      const SettingsScreen(),
    ];

    return Scaffold(
      body: Stack(
        children: [
          Positioned.fill(
            child: IndexedStack(index: _index, children: screens),
          ),
          // In-app banners for addNotification() calls app-wide, including
          // foreground push notifications — HomeShell stays mounted across
          // tab switches, so this renders regardless of which tab is active.
          StreamBuilder<List<HalendarNotification>>(
            stream: AuthService.instance.notificationsStream,
            initialData: AuthService.instance.notifications,
            builder: (context, snapshot) {
              final notifications = snapshot.data ?? [];
              return Align(
                alignment: Alignment.bottomCenter,
                child: SafeArea(
                  child: AnimatedList(
                    shrinkWrap: true,
                    key: AuthService.instance.listKey,
                    physics: const NeverScrollableScrollPhysics(),
                    initialItemCount: notifications.length,
                    itemBuilder: (context, index, animation) {
                      if (index >= notifications.length) {
                        return const SizedBox();
                      }
                      return buildNotificationCard(
                        notifications[index],
                        context,
                        animation,
                      );
                    },
                  ),
                ),
              );
            },
          ),
        ],
      ),
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
