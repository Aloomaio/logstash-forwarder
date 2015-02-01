;...

!define _VERSION "1.2.3.4"
!define _PRODUCT "Logstash Forwarder"
;...
;from {NSISDIR}\Examples\VersionInfo.nsi
;VIProductVersion "${_VERSION}"
;VIAddVersionKey /LANG=${LANG_ENGLISH} "ProductName" "${_PRODUCT}"
;VIAddVersionKey /LANG=${LANG_ENGLISH} "CompanyName" "SPS Commerce"
;VIAddVersionKey /LANG=${LANG_ENGLISH} "FileVersion" "${_VERSION}"
;VIAddVersionKey /LANG=${LANG_ENGLISH} "InternalName" "FileSetup.exe"
;...

Name "Logstash-Forwarder"
Outfile "logstash-forwarder-setup.exe"
InstallDir $PROGRAMFILES64\Logstash-Forwarder
RequestExecutionLevel admin

Section "Installer"
  ; Set the dir and move the file there
  SetOutPath $INSTDIR
  File logstash-forwarder.exe
  File logstash-forwarder.conf
  File nssm.exe

  ; Install and setup the service
  ExecWait '$INSTDIR\nssm.exe. install Logstash-Forwarder "$INSTDIR\logstash-forwarder.exe" "-config=logstash-forwarder.conf"'
  WriteRegStr HKLM "System\CurrentControlSet\Services\Logstash-Forwarder\Parameters" "AppStdout" "C:\logs\logstash-forwarder\out.log"
  WriteRegStr HKLM "System\CurrentControlSet\Services\Logstash-Forwarder\Parameters" "AppStderr" "C:\logs\logstash-forwarder\error.log"
  WriteRegDWORD HKLM "System\CurrentControlSet\Services\Logstash-Forwarder\Parameters" "AppRotate" 1
  WriteRegDWORD HKLM "System\CurrentControlSet\Services\Logstash-Forwarder\Parameters" "AppRotateBytes" 52428800

  ;Create log directory
  CreateDirectory "C:\logs\logstash-forwarder"

  ; Write the installation path into the registry
  WriteRegStr HKLM SOFTWARE\Logstash-Forwarder "Install_Dir" "$INSTDIR"

  ; Write the uninstall keys for Windows
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Logstash-Forwarder" "DisplayName" "Logstash-Forwarder"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Logstash-Forwarder" "UninstallString" '"$INSTDIR\uninstall.exe"'
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Logstash-Forwarder" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Logstash-Forwarder" "NoRepair" 1
  WriteUninstaller "uninstall.exe"
SectionEnd

Section "un.Uninstaller Section"
  ; Remove service
  ExecWait '"sc.exe" stop Logstash-Forwarder'
  ExecWait "$INSTDIR\nssm.exe remove Logstash-Forwarder confirm"

  ; Remove registry keys
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Logstash-Forwarder"
  DeleteRegKey HKLM SOFTWARE\Logstash-Forwarder

  ; Remove files and uninstaller
  Delete $INSTDIR\logstash-forwarder.conf
  Delete $INSTDIR\logstash-forwarder.exe
  Delete $INSTDIR\uninstall.exe
  Delete $INSTDIR\nssm.exe
  Delete $INSTDIR\.logstash-forwarder
  Delete $INSTDIR\.logstash-forwarder.old

  ; Remove directories used
  RMDir "$INSTDIR"
SectionEnd