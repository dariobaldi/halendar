import 'package:flutter/material.dart';

class MyButton extends StatelessWidget {
  final String text;
  final double width;
  final void Function(BuildContext) onTap;
  
  const MyButton({
    super.key,
    required this.text,
    required this.onTap,
    this.width = 300,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {onTap(context);},
      child: Container(
        width: width,
        padding: const EdgeInsets.all(15),
        margin: const EdgeInsets.symmetric(horizontal: 20),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.primary,
          borderRadius: BorderRadius.circular(10),
        ),
        child: Center(
          child: Text(
            text,
            style: TextStyle(color: Theme.of(context).colorScheme.onPrimary, fontSize: 19,),
          ),
        ),
      ),
    );
  }
}
 