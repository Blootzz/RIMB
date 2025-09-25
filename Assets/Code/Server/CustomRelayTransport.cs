using System;
using UnityEngine;
using Unity.Netcode;

public class CustomRelayTransport : NetworkTransport
{

    [Tooltip("Topology mode the transport will try to initialize with.")]
    [SerializeField]
    public NetworkTopologyTypes TopologyMode = NetworkTopologyTypes.ClientServer;
    private NetworkTopologyTypes ActiveTopologyMode = NetworkTopologyTypes.ClientServer;
    public override ulong ServerClientId => throw new NotImplementedException();

    [Serializable]
    public struct ConnectionAddressData
    {
        [Tooltip("IP address of the relay server (address to which clients will connect to).")]
        [SerializeField]
        public string RelayServerAddress;

        [Tooltip("UDP port of the server.")]
        [SerializeField]
        public ushort Port;
    }

    private static ConnectionAddressData DefaultConnectionAddressData = new ConnectionAddressData
    {
        RelayServerAddress = "127.0.0.1",
        Port = 7777
    };

    public ConnectionAddressData ConnectionData = DefaultConnectionAddressData;

    public override void DisconnectLocalClient()
    {
        throw new NotImplementedException();
    }

    public override void DisconnectRemoteClient(ulong clientId)
    {
        throw new NotImplementedException();
    }

    public override ulong GetCurrentRtt(ulong clientId)
    {
        throw new NotImplementedException();
    }

    public override void Initialize(NetworkManager networkManager = null)
    {
        throw new NotImplementedException();
    }

    public override NetworkEvent PollEvent(out ulong clientId, out ArraySegment<byte> payload, out float receiveTime)
    {
        throw new NotImplementedException();
    }

    public override void Send(ulong clientId, ArraySegment<byte> payload, NetworkDelivery networkDelivery)
    {
        throw new NotImplementedException();
    }

    public override void Shutdown()
    {
        throw new NotImplementedException();
    }

    public override bool StartClient()
    {
        throw new NotImplementedException();
    }

    public override bool StartServer()
    {
        throw new NotImplementedException();
    }
}