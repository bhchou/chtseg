package jchtseg;

import com.sun.jna.Memory;
import com.sun.jna.Native;
import com.sun.jna.Pointer;
import com.sun.jna.Library;
import com.sun.jna.Structure;
import java.util.Arrays;
import java.util.List;

import java.io.File;


public class jchtseg {

    private static String LIB_PATH = new File("jchtseg.so").getAbsolutePath();

    static ChtsegLib INSTANCE = (ChtsegLib) Native.loadLibrary(LIB_PATH, ChtsegLib.class);

    

    public static void main(String[] args) {
        returnChtseg( "sqlite3", "../db/chtseg.db", "現貨附發票 Raspberry Pi 樹莓派專用 USB電腦遙控器 帶無線鼠標無線鍵盤功能 萬能PC/紅外線遙控器");
    }


    private static void returnChtseg(String db, String conn, String test) {
        ChtsegResult segRet = INSTANCE.Getchtseg(db, conn, test);
        if ( segRet.r1.length() == 0 ) 
            System.out.println( "json result = " + segRet.r0 );
        else
            System.out.println( segRet.r1);
    }
}

public class ChtsegResult extends Structure implements Structure.ByValue {
    public String r0;
    public String r1;

    protected List<String> getFieldOrder() {
        return Arrays.asList("r0", "r1");
    }
}

public interface ChtsegLib extends Library {

    ChtsegResult Getchtseg(String db, String conn, String teststr);

}
