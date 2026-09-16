import 'package:flutter/material.dart';

class MyContainer extends StatelessWidget {
  final Widget child;
  final double? width;
  final double? height;
  final Color? color;
  const MyContainer({
    super.key,
    required this.child,
    this.width,
    this.height,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
          width: width,
          height: height,
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(12),
              boxShadow: const [
                BoxShadow(
                  color: Colors.black45,
                  offset: Offset(5, 5),
                  blurRadius: 10,
                ),
                BoxShadow(
                  color: Colors.white10,
                  offset: Offset(-5, -5),
                  blurRadius: 10,
                ),
              ]),
          child: child),
    );
  }
}

class MyCard extends StatelessWidget {
  final Widget child;
  final double? width;
  const MyCard({super.key, required this.child, this.width});

  @override
  Widget build(BuildContext context) {
    return Container(
        width: width,
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.primary.withAlpha(95),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: child,
        ),
      );
  }
}
