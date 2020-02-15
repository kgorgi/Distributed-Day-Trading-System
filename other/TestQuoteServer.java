import java.io.*;
import java.net.*;
import java.lang.*;
import java.util.Random;

public class TestQuoteServer {
    private static int requestCount = 1;
    private static int threadCount = 52;

    public static void main(String[] args) throws Exception {
        // Setup Threads
        QuoteTester[] testers = new QuoteTester[threadCount];
        Thread[] threads = new Thread[threadCount];

        for (int i = 0; i < threadCount; i++) {
            testers[i] = new QuoteTester(Integer.toString(1000000000 + i), requestCount);
            threads[i] = new Thread(testers[i]);
            threads[i].start();
        }

        // Wait for Threads to Complete
        for (int i = 0; i < threadCount; i++) {
            threads[i].join();
        }

        // Write Output
        String str = ReponsesToLog(testers);
        BufferedWriter writer = new BufferedWriter(new FileWriter("logfile.xml"));
        writer.write(str);
        writer.close();
    }

    private static String ReponsesToLog(QuoteTester[] testers) {
        StringBuilder str = new StringBuilder();
        str.append("<?xml version=\"1.0\"?>\n");
        str.append("<log>\n");

        for (int i = 0; i < testers.length; i++) {
            QuoteTester t = testers[i];
            for (int j = 0; j < t.responses.length; j++) {
                str.append("\t<quoteServer>\n");

                String[] data = t.responses[j].split(",");
                printTag(str, "timestamp", data[3]);
                printTag(str, "server", "quoteTest");
                printTag(str, "transactionNum", "70");
                printTag(str, "price", data[0]);
                printTag(str, "stockSymbol", data[1]);
                printTag(str, "username", data[2]);
                printTag(str, "quoteServerTime", data[3]);
                printTag(str, "cryptokey", data[4].trim());

                str.append("\t</quoteServer>\n");
            }
        }

        str.append("</log>\n");
        return str.toString();
    }

    private static void printTag(StringBuilder str, String tag, String value) {
        str.append("\t\t<");
        str.append(tag);
        str.append(">");
        str.append(value);
        str.append("</");
        str.append(tag);
        str.append(">\n");
    }
}

class QuoteTester implements Runnable {
    private static String alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    private static String quoteServerAddress = "192.168.1.100";
    private static int quoteServerPort = 4443;
    private Random rand = new Random();

    public String[] responses;
    private int requestCount;
    private String userid;

    public QuoteTester(String userid, int requestCount) {
        this.responses = new String[requestCount];
        this.requestCount = requestCount;
        this.userid = userid;
    }

    public void run() {
        try {
            for (int i = 0; i < requestCount; i++) {
                Socket kkSocket = new Socket(quoteServerAddress, quoteServerPort);
                PrintWriter out = new PrintWriter(kkSocket.getOutputStream(), true);
                BufferedReader in = new BufferedReader(new InputStreamReader(kkSocket.getInputStream()));

                String fromUser = createStock() + "," + this.userid + "\n";
                out.println(fromUser);

                String fromServer = in.readLine();

                this.responses[i] = fromServer;

                out.close();
                in.close();
                kkSocket.close();
            }
        } catch (Exception ex) {
            System.err.println(ex);
            System.exit(0);
        }
    }

    private String createStock() {
        return "" + alphabet.charAt(rand.nextInt(26)) + alphabet.charAt(rand.nextInt(26))
                + alphabet.charAt(rand.nextInt(26));
    }
}