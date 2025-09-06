using Unity.Netcode;
using Unity.VisualScripting;
using UnityEditor;
using UnityEngine;
using UnityEngine.UIElements;

public class NetTestUI : MonoBehaviour
{

    private UIDocument document;
    private Label colorPreviewLbl;
    private Slider rSlider;
    private Slider gSlider;
    private Slider bSlider;
    private Button connectBtn;
    private Button disconnectBtn;
    private Button startHostBtn;

    void OnEnable()
    {
        document = GetComponent<UIDocument>();
        VisualElement root = document.rootVisualElement;
        colorPreviewLbl = root.Q<Label>("color_preview");
        rSlider = root.Q<Slider>("r_slider");
        gSlider = root.Q<Slider>("g_slider");
        bSlider = root.Q<Slider>("b_slider");
        connectBtn = root.Q<Button>("connect");
        disconnectBtn = root.Q<Button>("disconnect");
        startHostBtn = root.Q<Button>("start_host");
        UpdateColorPreview();
        rSlider.RegisterValueChangedCallback(OnSliderUpdate);
        gSlider.RegisterValueChangedCallback(OnSliderUpdate);
        bSlider.RegisterValueChangedCallback(OnSliderUpdate);
        connectBtn.clicked += OnConnect;
        disconnectBtn.clicked += OnDisconnect;
        startHostBtn.clicked += OnStartHost;
        NetworkManager nm = NetworkManager.Singleton;
        if (nm != null)
            NetworkManager.Singleton.OnPreShutdown += OnDisconnect; // Will happen automatically on clients if the server shuts down.
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

    void OnDisable()
    {
        connectBtn.clicked -= OnConnect;
        disconnectBtn.clicked -= OnDisconnect;
        startHostBtn.clicked -= OnStartHost;
        rSlider.UnregisterValueChangedCallback(OnSliderUpdate);
        gSlider.UnregisterValueChangedCallback(OnSliderUpdate);
        bSlider.UnregisterValueChangedCallback(OnSliderUpdate);
        NetworkManager nm = NetworkManager.Singleton;
        if (nm != null)
            NetworkManager.Singleton.OnPreShutdown -= OnDisconnect;
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
        if (nm.IsClient || (!nm.IsListening && nm.StartClient()))
            DisplayConnected("Disconnect");
        else if (nm.IsHost)
            DisplayConnected("Stop Host");
        else
            ResetDisplay();
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
        else
            ResetDisplay();
    }

}
