using Unity.Netcode;
using UnityEngine;
using Unity.Netcode.Transports.UTP;

public class NMAutostartServer : MonoBehaviour
{
    void Start()
    {
#if UNITY_SERVER
        Application.targetFrameRate = (int)(1f / Time.fixedDeltaTime); // Prevent server from shitting itself with 1000 updates per second.
        NetworkManager nm = NetworkManager.Singleton;
        if (nm == null)
            return;
        UnityTransport transport = nm.GetComponent<UnityTransport>();
        if (transport == null)
            return;
        transport.SetConnectionData("0.0.0.0", transport.ConnectionData.Port);
        nm.StartServer();
        print($"Server is listening at address {transport.ConnectionData.Address}:{transport.ConnectionData.Port}");
#endif
    }
}
