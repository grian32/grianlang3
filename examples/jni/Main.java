public class Main {

    // NOTE: you've gotta make in the root directory before you can run this
    public native int add(int a, int b);

    static {
        System.loadLibrary("add");
    }

    public static void main(String[] args) {
        Main m = new Main();
        System.out.printf("Hello, 2+2=%d\n", m.add(2, 2));
    }
}
