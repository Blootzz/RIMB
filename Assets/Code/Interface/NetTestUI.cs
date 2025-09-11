using Unity.Netcode;
using UnityEngine;
using UnityEngine.UIElements;
using UnityEngine.InputSystem;
using Unity.Netcode.Transports.UTP;
using UnityEngine.TextCore;
using Netcode.Transports.Facepunch;

public class NetTestUI : MonoBehaviour
{

    private UIDocument document;
    private Label colorPreviewLbl;
    private Slider rSlider;
    private Slider gSlider;
    private Slider bSlider;
    private Slider sensitivitySlider;
    private Button connectBtn;
    private Button disconnectBtn;
    private Button startHostBtn;
    private TextField ipAddressField;

    void OnEnable()
    {
        document = GetComponent<UIDocument>();
        VisualElement root = document.rootVisualElement;
        colorPreviewLbl = root.Q<Label>("color_preview");
        rSlider = root.Q<Slider>("r_slider");
        gSlider = root.Q<Slider>("g_slider");
        bSlider = root.Q<Slider>("b_slider");
        sensitivitySlider = root.Q<Slider>("sensitivity");
        connectBtn = root.Q<Button>("connect");
        disconnectBtn = root.Q<Button>("disconnect");
        startHostBtn = root.Q<Button>("start_host");
        ipAddressField = root.Q<TextField>("server_address");
        UpdateColorPreview();
        rSlider.RegisterValueChangedCallback(OnSliderUpdate);
        gSlider.RegisterValueChangedCallback(OnSliderUpdate);
        bSlider.RegisterValueChangedCallback(OnSliderUpdate);
        sensitivitySlider.RegisterValueChangedCallback(UpdateSensitivity);
        connectBtn.clicked += OnConnect;
        disconnectBtn.clicked += OnDisconnect;
        startHostBtn.clicked += OnStartHost;
        NetworkManager nm = NetworkManager.Singleton;
        if (nm != null)
            NetworkManager.Singleton.OnPreShutdown += ResetDisplay; // Will happen automatically on clients if the server shuts down.

        // Interface locks
        InputAction mouseLockToggle = InputSystem.actions.FindAction("MouseLockToggle");
        if (mouseLockToggle != null)
            mouseLockToggle.performed += ToggleMouseLock;

        // Transport host field setup
        NetworkTransport transport = nm.GetComponent<NetworkTransport>();
        if (transport != null)
        {
            if (transport is UnityTransport)
            {
                // Standard UTP direct connect transport
                ipAddressField.label = "IP Address";
            }
            else if (transport is FacepunchTransport)
            {
                // Steamworks transport
                ipAddressField.label = "Target Steam ID";
            }
        }
    }

    void OnDisable()
    {
        connectBtn.clicked -= OnConnect;
        disconnectBtn.clicked -= OnDisconnect;
        startHostBtn.clicked -= OnStartHost;
        rSlider.UnregisterValueChangedCallback(OnSliderUpdate);
        gSlider.UnregisterValueChangedCallback(OnSliderUpdate);
        bSlider.UnregisterValueChangedCallback(OnSliderUpdate);
        sensitivitySlider.UnregisterValueChangedCallback(UpdateSensitivity);
        NetworkManager nm = NetworkManager.Singleton;
        if (nm != null)
            NetworkManager.Singleton.OnPreShutdown -= ResetDisplay;
    }

    void ToggleMouseLock(InputAction.CallbackContext context)
    {
        if (UnityEngine.Cursor.lockState == CursorLockMode.None)
            UnityEngine.Cursor.lockState = CursorLockMode.Locked;
        else
            UnityEngine.Cursor.lockState = CursorLockMode.None;
    }

    void UpdateSensitivity(ChangeEvent<float> e)
    {
        CapManControl[] players = FindObjectsByType<CapManControl>(FindObjectsSortMode.None);
        foreach (CapManControl player in players)
        {
            player.LookSensitivity = e.newValue;
        }
    }

    void UpdateColorPreview()
    {
        float r = rSlider.value;
        float g = gSlider.value;
        float b = bSlider.value;
        colorPreviewLbl.style.color = new(new Color(1 - r, 1 - g, 1 - b));
        colorPreviewLbl.style.backgroundColor = new(new Color(r, g, b));
        CapManControl[] players = FindObjectsByType<CapManControl>(FindObjectsSortMode.None);
        foreach (CapManControl player in players) {
            player.PlayerColor = new Color(r, g, b); // Non-owner writes are ignored
        }
    }

    void OnSliderUpdate(ChangeEvent<float> e)
    {
        UpdateColorPreview();
    }

    void ResetDisplay()
    {
        disconnectBtn.style.display = DisplayStyle.None;
        connectBtn.style.display = DisplayStyle.Flex;
        startHostBtn.style.display = DisplayStyle.Flex;
    }

    void DisplayConnected(string disconnectText)
    {
        connectBtn.style.display = DisplayStyle.None;
        startHostBtn.style.display = DisplayStyle.None;
        disconnectBtn.text = disconnectText;
        disconnectBtn.style.display = DisplayStyle.Flex;
    }

    void OnConnect()
    {
        NetworkManager nm = NetworkManager.Singleton;
        if (nm == null)
            return;
        NetworkTransport transport = nm.GetComponent<NetworkTransport>();
        if (!nm.IsListening)
        {
            if (transport is UnityTransport)
            {
                UnityTransport ut = (UnityTransport)transport;
                ut.SetConnectionData(ipAddressField.value, ut.ConnectionData.Port);
            }
            else if (transport is FacepunchTransport)
            {
                FacepunchTransport ft = (FacepunchTransport)transport;
                ft.targetSteamId = ulong.Parse(ipAddressField.value);
            }
        }
        if (nm.IsClient || (!nm.IsListening && nm.StartClient()))
                DisplayConnected("Disconnect");
            else if (nm.IsHost)
                DisplayConnected("Stop Host");
    }

    void OnDisconnect()
    {
        ResetDisplay();
        NetworkManager nm = NetworkManager.Singleton;
        if (nm != null && nm.IsListening)
            nm.Shutdown();
    }

    void OnStartHost()
    {
        NetworkManager nm = NetworkManager.Singleton;
        if (nm == null)
            return;
        if (nm.IsHost || (!nm.IsListening && nm.StartHost()))
            DisplayConnected("Stop Host");
        else if (nm.IsClient)
            DisplayConnected("Disconnect");
    }

}
