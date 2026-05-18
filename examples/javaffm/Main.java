import java.lang.foreign.*;
import java.lang.invoke.MethodHandle;

public class Main {
    // NOTE: you've gotta make in the root directory before you can run this
    static int invokeNativeAdd(int a, int b) throws Throwable {
        System.loadLibrary("add");

        SymbolLookup lookup = SymbolLookup.loaderLookup();

        FunctionDescriptor desc = FunctionDescriptor.of(
            ValueLayout.JAVA_INT,
            ValueLayout.JAVA_INT,
            ValueLayout.JAVA_INT
        );

        MethodHandle add = Linker.nativeLinker().downcallHandle(
            lookup.find("add").orElseThrow(),
            desc
        );

        return (int) add.invoke(a, b);
    }


    public static void main(String[] args) throws Throwable {
        System.out.printf("Hello, 2+2=%d\n", Main.invokeNativeAdd(2, 2));
    }
}
