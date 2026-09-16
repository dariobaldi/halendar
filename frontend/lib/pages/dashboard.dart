import 'package:halendar_front/components/nav_bar.dart';
import 'package:halendar_front/components/responsive.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:halendar_front/services/notifications.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class DashboardScreen extends StatelessWidget {
  final Widget child;
  const DashboardScreen({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Scaffold(
        backgroundColor: Theme.of(context).colorScheme.surface,
        drawer: const NavBar(),
        body: Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (Responsive.isDesktop(context)) const NavBar(),
            (AuthService.instance.accessLevel < 5 && !isWorkingHours())
                ? Expanded(child: OutsideWorkingHoursPage())
                : Expanded(
                    child: Stack(
                      children: [
                        Positioned.fill(child: child),
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
                                  reverse: false,
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
                  ),
          ],
        ),
      ),
    );
  }
}

class TopBar extends StatefulWidget {
  const TopBar({super.key});

  @override
  State<TopBar> createState() => _TopBarState();
}

class _TopBarState extends State<TopBar> {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(8.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          if (!Responsive.isDesktop(context))
            IconButton(
              icon: const Icon(Icons.menu),
              onPressed: () {
                Scaffold.of(context).openDrawer();
              },
            ),
          const SizedBox(width: 15),
          // if (!Responsive.isMobile(context)) const SearchWidget(),
          // const SizedBox(width: 15),
          // Row(
          //   children: [
          //     IconButton(onPressed: () {}, icon: const Icon(Icons.message)),
          //     IconButton(onPressed: () {}, icon: const Icon(Icons.settings)),
          //     const SizedBox(width: 10),
          //   ],
          // ),
          // IconButton(
          //     onPressed: () {
          //       AuthService.instance.logOut();
          //     },
          //     icon: const Icon(Icons.logout)),
        ],
      ),
    );
  }
}

class SearchWidget extends StatelessWidget {
  const SearchWidget({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: Responsive.isDesktop(context) ? 500 : 380,
      child: TextField(
        onChanged: (value) {},
        decoration: const InputDecoration(
          hintText: "Recherche",
          filled: true,
          suffixIcon: Padding(
            padding: EdgeInsets.all(0.75), //15
            child: Icon(Icons.search),
          ),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.all(Radius.circular(10)),
            borderSide: BorderSide.none,
          ),
        ),
      ),
    );
  }
}

class OutsideWorkingHoursPage extends StatelessWidget {
  const OutsideWorkingHoursPage({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      body: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [colorScheme.surface, colorScheme.surfaceContainerHighest],
          ),
        ),
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(32),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 500),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  // Illustration
                  Container(
                    width: 140,
                    height: 140,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: colorScheme.primaryContainer,
                      boxShadow: [
                        BoxShadow(
                          color: colorScheme.shadow.withValues(alpha: 0.08),
                          blurRadius: 30,
                          spreadRadius: 5,
                        ),
                      ],
                    ),
                    child: Stack(
                      alignment: Alignment.center,
                      children: [
                        Icon(
                          Icons.nightlight_round,
                          size: 64,
                          color: colorScheme.primary,
                        ),
                        Positioned(
                          right: 24,
                          top: 25,
                          child: Icon(
                            Icons.star_rounded,
                            size: 22,
                            color: colorScheme.primary,
                          ),
                        ),
                        Positioned(
                          left: 25,
                          bottom: 28,
                          child: Icon(
                            Icons.star_rounded,
                            size: 14,
                            color: colorScheme.primary,
                          ),
                        ),
                      ],
                    ),
                  ),

                  const SizedBox(height: 16),

                  // Main Title
                  Text(
                    "Halendar n'est pas disponible en dehors des heures de travail",
                    textAlign: TextAlign.center,
                    style: theme.textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w700,
                      height: 1.3,
                      color: colorScheme.onSurface,
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Supportive Subtitle
                  Text(
                    'Profitez de votre temps libre pour vous reposer :)',
                    textAlign: TextAlign.center,
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                      height: 1.5,
                    ),
                  ),

                  if (DateTime.now().weekday != DateTime.friday &&
                      DateTime.now().weekday != DateTime.saturday)
                    const SizedBox(height: 40),

                  if (DateTime.now().weekday != DateTime.friday &&
                      DateTime.now().weekday != DateTime.saturday)
                    Text(
                      'À demain 👋',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: colorScheme.onSurface,
                      ),
                    ),

                  const SizedBox(height: 32),

                  // Working hours
                  // Container(
                  //   padding: const EdgeInsets.symmetric(
                  //     horizontal: 20,
                  //     vertical: 14,
                  //   ),
                  //   decoration: BoxDecoration(
                  //     color: colorScheme.surface.withValues(alpha: 0.7),
                  //     borderRadius: BorderRadius.circular(16),
                  //     border: Border.all(
                  //       color: colorScheme.outlineVariant,
                  //     ),
                  //   ),
                  //   child: Row(
                  //     mainAxisSize: MainAxisSize.min,
                  //     children: [
                  //       Icon(
                  //         Icons.schedule_rounded,
                  //         size: 20,
                  //         color: colorScheme.primary,
                  //       ),
                  //       const SizedBox(width: 10),
                  //       Text(
                  //         'Lundi – Vendredi  •  08:00 – 18:00',
                  //         style: theme.textTheme.bodyMedium?.copyWith(
                  //           fontWeight: FontWeight.w600,
                  //           color: colorScheme.onSurfaceVariant,
                  //         ),
                  //       ),
                  //     ],
                  //   ),
                  // ),
                  ActionChip(
                    avatar: const Icon(Icons.logout),
                    label: Text("Se déconnecter", style: GoogleFonts.ubuntu()),
                    onPressed: () {
                      AuthService.instance.logOut();
                    },
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
