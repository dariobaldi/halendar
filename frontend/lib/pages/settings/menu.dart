import 'package:halendar_front/pages/dashboard.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:url_launcher/url_launcher.dart';

class SettingsMenu extends StatelessWidget {
  const SettingsMenu({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(5),
      children: [
        TopBar(),
        if ((AuthService.instance.user?.accessLevel ?? 0) >= 10)
          MenuButton(
            title: "Utilisateurs",
            icon: Icons.groups,
            url: "/settings/users",
          ),
        MenuButton(
          title: "Mot de Passe",
          icon: Icons.key,
          url: "/settings/change_password",
        ),
        MenuButton(
          title: "Marketplaces",
          icon: Icons.store,
          url: "/settings/marketplaces",
        ),
        MenuButtonLink(
          title: "Android",
          icon: Icons.android,
          link:
              "https://www.dropbox.com/scl/fo/shuvjrev09en11qxn92l4/AKR8aps0xwk5ryF63gS1WTY?rlkey=9l3dcbarkkazv47n3bq3snbi6&st=aiu3k8ec&dl=0",
        ),
      ],
    );
  }
}

class MenuButton extends StatelessWidget {
  final String title;
  final IconData icon;
  final String url;
  final String imageUrl;
  const MenuButton({
    super.key,
    required this.title,
    this.icon = Icons.question_mark,
    required this.url,
    this.imageUrl = "",
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 5),
      child: Center(
        child: SizedBox(
          width: 250,
          height: 50,
          child: OutlinedButton(
            onPressed: () {
              context.go(url);
            },
            child: Row(
              children: [
                (imageUrl == "")
                    ? Icon(icon)
                    : Image.asset(imageUrl, height: 25),
                const SizedBox(width: 15),
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class MenuButtonLink extends StatelessWidget {
  final String title;
  final IconData icon;
  final String link;
  final String imageUrl;

  const MenuButtonLink({
    super.key,
    required this.title,
    this.icon = Icons.question_mark,
    required this.link,
    this.imageUrl = "",
  });

  Future<void> openInvoice() async {
    final Uri url = Uri.parse(link);
    if (!await launchUrl(url, webOnlyWindowName: "_blank")) {
      throw Exception('Could not launch tracking URL');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 5),
      child: Center(
        child: SizedBox(
          width: 250,
          height: 50,
          child: FilledButton(
            onPressed: () async {
              openInvoice();
            },
            child: Row(
              children: [
                (imageUrl == "")
                    ? Icon(icon)
                    : Image.asset(imageUrl, height: 25),
                const SizedBox(width: 15),
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
