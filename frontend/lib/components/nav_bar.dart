import 'package:halendar_front/components/responsive.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';

class NavBar extends StatelessWidget {
  const NavBar({super.key});

  @override
  Widget build(BuildContext context) {
    return Drawer(
      backgroundColor: Theme.of(context).colorScheme.surface,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: ListView(
          children: [
            DrawerHeader(
              child: Stack(
                children: [
                  Column(
                    children: [
                      GestureDetector(
                        onTap: () {
                          context.go("/");
                          Scaffold.of(context).closeDrawer();
                        },
                        child: Image.asset(
                          "./lib/images/logo/ic_launcher.png",
                          width: 200,
                          height: 80,
                        ),
                      ),
                      Row(
                        children: [
                          IconButton(
                            onPressed: () {
                              context.go(Uri(path: "/settings/").toString());
                              Scaffold.of(context).closeDrawer();
                            },
                            icon: const Icon(Icons.settings),
                          ),
                          GreetingMessage(),
                        ],
                      ),
                    ],
                  ),
                  if (!Responsive.isDesktop(context))
                    Positioned(
                      right: 0,
                      child: IconButton(
                        onPressed: () {
                          Scaffold.of(context).closeDrawer();
                        },
                        icon: const Icon(Icons.arrow_back),
                      ),
                    ),
                ],
              ),
            ),
            const SizedBox(height: 10),
            const MyDrawerItem(title: "Acceuil", icon: Icons.home),
            const MyDrawerItem(
              title: "Listes",
              icon: Icons.list,
              location: '/list',
            ),
            const MyDrawerItem(
              title: "Tâches",
              icon: Icons.task_alt,
              location: '/tasks',
            ),
            const SizedBox(height: 10),
            const MyDrawerItem(
              title: "Picking",
              icon: Icons.category,
              location: '/picking',
            ),
            const MyDrawerItem(
              title: "Colisage",
              icon: Icons.inventory,
              location: '/packing',
            ),
            const MyDrawerItem(
              title: "Envois",
              icon: Icons.local_shipping,
              location: '/shipping',
            ),
            const MyDrawerItem(
              title: "FBA",
              icon: Icons.warehouse,
              location: '/fba',
            ),
            const SizedBox(height: 15),
            const MyDrawerItem(
              title: "Produits",
              icon: Icons.shelves,
              location: '/products',
            ),
            const MyDrawerItem(
              title: "Restocker",
              icon: Icons.move_down,
              location: '/restocking',
            ),
            const MyDrawerItem(
              title: "Stock",
              icon: Icons.forklift,
              location: '/stock_management',
            ),
            if (AuthService.instance.accessLevel >= 10)
              const SizedBox(height: 15),
            if (AuthService.instance.accessLevel >= 10)
              const MyDrawerItem(
                title: "Stats",
                icon: Icons.query_stats,
                location: '/stats',
              ),
            if (AuthService.instance.accessLevel >= 100)
              const MyDrawerItem(
                title: "Test",
                icon: Icons.bug_report,
                location: '/test',
              ),
            const SizedBox(height: 15),
            if (AuthService.instance.accessLevel >= 100)
              ElevatedButton(
                onPressed: () {
                  addNotification(
                    title: "Error",
                    content: "ebay API error [5]: Erreur d'analyse XML.",
                    imageUrl:
                        "https://cdn.shopify.com/s/files/1/0117/0859/6305/products/234056584928-0.jpg?v=1624044902",
                    type: "error",
                  );
                },
                child: Text("Test Notification"),
              ),
            TextButton(
              onPressed: () {
                AuthService.instance.logOut();
              },
              child: ListTile(
                leading: const Icon(Icons.logout),
                title: Text("Se déconnecter", style: GoogleFonts.ubuntu()),
              ),
            ),
            const SizedBox(height: 15),
          ],
        ),
      ),
    );
  }
}

class MyDrawerItem extends StatelessWidget {
  final String title;
  final IconData icon;
  final String? location;

  const MyDrawerItem({
    super.key,
    required this.title,
    this.icon = Icons.question_answer,
    this.location,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.0),
      child: FilledButton.tonal(
        onPressed: () {
          context.go(location ?? "/");
          Scaffold.of(context).closeDrawer();
        },
        child: ListTile(
          leading: Icon(icon),
          title: Text(title, style: GoogleFonts.ubuntu(fontSize: 18)),
        ),
      ),
    );
  }
}

class GreetingMessage extends StatelessWidget {
  final AuthUser _user = AuthService.instance.user ?? anonymousUser;
  GreetingMessage({super.key});

  String getGreeting() {
    DateTime now = DateTime.now();
    int hour = now.hour;
    int weekday = now.weekday;

    // Weekend check: Saturday (7) and Sunday (6)
    if (weekday == DateTime.saturday || weekday == DateTime.sunday) {
      return "Bon week-end, ";
    }

    // Weekday (Monday to Friday) logic
    if (hour >= 6 && hour < 12) {
      return "Bonjour, ";
    } else if (hour >= 12 && hour < 14) {
      return "Bon appétit, ";
    } else if (hour >= 14 && hour < 18) {
      return "Salut, ";
    } else if (hour >= 18 && hour < 22) {
      return "Bonne soirée, ";
    } else {
      return "Bonne nuit, ";
    }
  }

  @override
  Widget build(BuildContext context) {
    return Text(
      getGreeting() + _user.name,
      style: GoogleFonts.ubuntu(textStyle: const TextStyle(fontSize: 18)),
    );
  }
}
