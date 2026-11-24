package id.warkop.pos;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;

/** Stasiun cetak menyala lagi sendiri setelah HP/tablet dinyalakan ulang. */
public class BootReceiver extends BroadcastReceiver {
    @Override
    public void onReceive(Context context, Intent intent) {
        if (Intent.ACTION_BOOT_COMPLETED.equals(intent.getAction())) StasiunService.terapkan(context);
    }
}
