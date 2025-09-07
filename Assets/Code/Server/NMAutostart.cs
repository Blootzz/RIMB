using Unity.Netcode;
using UnityEngine;
using Unity.Netcode.Transports.UTP;

public class NMAutostartServer : MonoBehaviour
{
    void Start()
    {
#if UNITY_SERVER
        NetworkManager nm = NetworkManager.Singleton;
        if (nm == null)
            return;
        UnityTransport transport = nm.GetComponent<UnityTransport>();
        if (transport == null)
            return;
        nm.StartServer();
        print($"Server is listening at address {transport.ConnectionData.Address}:{transport.ConnectionData.Port}");
#endif
    }
}
